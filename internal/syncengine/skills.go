package syncengine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// filepathRel is a seam over filepath.Rel so tests can inject the otherwise
// unreachable error case (mirrors engine.go's getwd seam).
var filepathRel = filepath.Rel

// iterSkillDirs returns every directory under root that directly contains
// SKILL.md, at any depth, without descending into a directory once matched.
// Mirrors sync_engine.py's _iter_skill_dirs.
func iterSkillDirs(root string) []string {
	if !exists(root) {
		return nil
	}
	if exists(filepath.Join(root, "SKILL.md")) {
		return []string{root}
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	sort.Strings(dirs)

	var result []string
	for _, name := range dirs {
		result = append(result, iterSkillDirs(filepath.Join(root, name))...)
	}
	return result
}

// collectSkills maps skill name -> directory across sources; later sources
// win on name collision. Mirrors sync_engine.py's _collect_skills.
func collectSkills(sources ...string) map[string]string {
	skills := map[string]string{}
	for _, source := range sources {
		for _, skillDir := range iterSkillDirs(source) {
			skills[filepath.Base(skillDir)] = skillDir
		}
	}
	return skills
}

// SyncFlattenedSkills flattens one or more (possibly grouped) skill source
// roots into dest, so every skill directory sits directly under dest.
// Mirrors sync_engine.py's sync_flattened_skills.
func SyncFlattenedSkills(dest string, sources ...string) error {
	desired := collectSkills(sources...)

	if isSymlink(dest) {
		if err := os.Remove(dest); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}

	entries, err := os.ReadDir(dest)
	if err != nil {
		return err
	}
	lastSource := sources[len(sources)-1]
	for _, entry := range entries {
		name := entry.Name()
		if _, ok := desired[name]; ok {
			continue
		}
		childPath := filepath.Join(dest, name)
		if isSymlink(childPath) {
			logger.Info(fmt.Sprintf("Removing stale skill symlink %s", childPath))
			if err := os.Remove(childPath); err != nil {
				return err
			}
			continue
		}
		target := filepath.Join(lastSource, name)
		if err := move(childPath, target); err != nil {
			return err
		}
		desired[name] = target
	}

	names := make([]string, 0, len(desired))
	for name := range desired {
		names = append(names, name)
	}
	sort.Strings(names) // deterministic order; final filesystem state is order-independent

	for _, name := range names {
		if err := symlink(desired[name], filepath.Join(dest, name)); err != nil {
			return err
		}
	}
	return nil
}

// SyncPartiallyGroupedSkills syncs flat skills from flattenSrc directly under
// dest, while skills under groupedSrc keep their relative group subfolder at
// dest. A grouped skill's name overrides a flat skill of the same name. A
// new unmanaged skill found inside an existing group folder at dest is
// absorbed back into that same group in groupedSrc; one found outside every
// known group is absorbed into flattenSrc instead, since it has no group.
func SyncPartiallyGroupedSkills(dest, flattenSrc, groupedSrc string) error {
	flat := collectSkills(flattenSrc)

	grouped := map[string]string{} // rel (slash-separated) -> absDir
	nameToRel := map[string]string{}
	groupDirs := map[string]struct{}{} // known intermediate group-folder rel paths
	for _, skillDir := range iterSkillDirs(groupedSrc) {
		rel, err := filepathRel(groupedSrc, skillDir)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		grouped[rel] = skillDir
		nameToRel[filepath.Base(skillDir)] = rel

		segments := strings.Split(rel, "/")
		for i := 1; i < len(segments); i++ {
			groupDirs[strings.Join(segments[:i], "/")] = struct{}{}
		}
	}

	for name := range nameToRel {
		delete(flat, name)
	}

	if isSymlink(dest) {
		if err := os.Remove(dest); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}

	if err := absorbGroupedSkills(dest, "", flattenSrc, groupedSrc, flat, grouped, groupDirs); err != nil {
		return err
	}

	flatNames := make([]string, 0, len(flat))
	for name := range flat {
		flatNames = append(flatNames, name)
	}
	sort.Strings(flatNames)
	for _, name := range flatNames {
		if err := symlink(flat[name], filepath.Join(dest, name)); err != nil {
			return err
		}
	}

	groupedRels := make([]string, 0, len(grouped))
	for rel := range grouped {
		groupedRels = append(groupedRels, rel)
	}
	sort.Strings(groupedRels)
	for _, rel := range groupedRels {
		destPath := filepath.Join(dest, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return err
		}
		if err := symlink(grouped[rel], destPath); err != nil {
			return err
		}
	}
	return nil
}

// containsSymlink reports whether dirPath, or any directory beneath it,
// contains a symlink.
func containsSymlink(dirPath string) (bool, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		childPath := filepath.Join(dirPath, entry.Name())
		if isSymlink(childPath) {
			return true, nil
		}
		if entry.IsDir() {
			found, err := containsSymlink(childPath)
			if err != nil {
				return false, err
			}
			if found {
				return true, nil
			}
		}
	}
	return false, nil
}

// absorbStaleGroupSubtree cleans up a dest-side directory whose relative
// path is no longer known to the repo: stale symlinks underneath it are
// deleted, remaining real content is moved back into srcRoot at the same
// relative path, and the directory itself is removed once empty.
func absorbStaleGroupSubtree(dirPath, rel, srcRoot string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	empty := true
	for _, entry := range entries {
		name := entry.Name()
		childPath := filepath.Join(dirPath, name)
		childRel := rel + "/" + name
		switch {
		case isSymlink(childPath):
			logger.Info(fmt.Sprintf("Removing stale skill symlink %s", childPath))
			if err := os.Remove(childPath); err != nil {
				return err
			}
		case entry.IsDir():
			if err := absorbStaleGroupSubtree(childPath, childRel, srcRoot); err != nil {
				return err
			}
			if exists(childPath) {
				empty = false
			}
		default:
			target := filepath.Join(srcRoot, filepath.FromSlash(childRel))
			if err := move(childPath, target); err != nil {
				return err
			}
			if exists(childPath) {
				empty = false
			}
		}
	}
	if empty {
		logger.Info(fmt.Sprintf("Removing stale group directory %s", dirPath))
		if err := os.RemoveAll(dirPath); err != nil {
			return err
		}
	}
	return nil
}

// absorbGroupedSkills moves unmanaged real entries under dirPath (dest, or
// one of its known group subdirectories, identified by relPrefix) into the
// repo: entries at the top level (relPrefix == "") go to flattenSrc, entries
// inside a known group go to groupedSrc at the same relative path. It
// recurses only through directories already known to be part of the group
// structure (groupDirs), so an entirely new group is absorbed as one unit
// into flattenSrc rather than merged into an existing group.
func absorbGroupedSkills(dirPath, relPrefix, flattenSrc, groupedSrc string, flat, grouped map[string]string, groupDirs map[string]struct{}) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		childPath := filepath.Join(dirPath, name)

		childRel := name
		if relPrefix != "" {
			childRel = relPrefix + "/" + name
		}

		if isSymlink(childPath) {
			var stale bool
			if relPrefix == "" {
				_, ok := flat[name]
				stale = !ok
			} else {
				_, ok := grouped[childRel]
				stale = !ok
			}
			if stale {
				logger.Info(fmt.Sprintf("Removing stale skill symlink %s", childPath))
				if err := os.Remove(childPath); err != nil {
					return err
				}
			}
			continue
		}

		if relPrefix == "" {
			if _, ok := flat[name]; ok {
				continue
			}
		}
		if _, ok := grouped[childRel]; ok {
			continue
		}
		if _, ok := groupDirs[childRel]; ok {
			if err := absorbGroupedSkills(childPath, childRel, flattenSrc, groupedSrc, flat, grouped, groupDirs); err != nil {
				return err
			}
			continue
		}

		srcRoot := flattenSrc
		if relPrefix != "" {
			srcRoot = groupedSrc
		}
		if entry.IsDir() {
			hasSymlink, err := containsSymlink(childPath)
			if err != nil {
				return err
			}
			if hasSymlink {
				if err := absorbStaleGroupSubtree(childPath, childRel, srcRoot); err != nil {
					return err
				}
				continue
			}
		}

		if relPrefix == "" {
			target := filepath.Join(flattenSrc, name)
			if err := move(childPath, target); err != nil {
				return err
			}
			flat[name] = target
			continue
		}
		target := filepath.Join(groupedSrc, filepath.FromSlash(childRel))
		if err := move(childPath, target); err != nil {
			return err
		}
		grouped[childRel] = target
	}
	return nil
}
