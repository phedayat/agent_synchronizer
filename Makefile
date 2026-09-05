clean:
	find . -type d -name '__pycache__' -prune -exec rm -rf {} +
	
	rm -rf build
	rm -rf dist
	rm -rf *.egg-info
	rm -rf .pytest_cache
	rm -rf .ruff_cache

test:
	uv run pytest -sv

lint:
	uvx ruff check --fix .

format:
	uvx ruff format .

typecheck:
	uvx ty check .

prepare: lint format typecheck test
