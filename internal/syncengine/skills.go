package syncengine

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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
// dest. A grouped skill's name overrides a flat skill of the same name.
func SyncPartiallyGroupedSkills(dest, flattenSrc, groupedSrc string) error {
	flat := collectSkills(flattenSrc)

	grouped := map[string]string{}
	nameToRel := map[string]string{}
	for _, skillDir := range iterSkillDirs(groupedSrc) {
		rel, err := filepath.Rel(groupedSrc, skillDir)
		if err != nil {
			return err
		}
		grouped[rel] = skillDir
		nameToRel[filepath.Base(skillDir)] = rel
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

	desiredTop := map[string]struct{}{}
	for name := range flat {
		desiredTop[name] = struct{}{}
	}
	for rel := range grouped {
		top := strings.Split(filepath.ToSlash(rel), "/")[0]
		desiredTop[top] = struct{}{}
	}

	entries, err := os.ReadDir(dest)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		if _, ok := desiredTop[name]; ok {
			continue
		}
		childPath := filepath.Join(dest, name)
		if isSymlink(childPath) {
			continue
		}
		target := filepath.Join(groupedSrc, name)
		if err := move(childPath, target); err != nil {
			return err
		}
		grouped[name] = target
		desiredTop[name] = struct{}{}
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
		destPath := filepath.Join(dest, rel)
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return err
		}
		if err := symlink(grouped[rel], destPath); err != nil {
			return err
		}
	}
	return nil
}
