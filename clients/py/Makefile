.PHONY: all
all: install lint test docs

.PHONY: install
install:
	pip install -e ".[dev]"
	pip install -r requirements-dev.txt

.PHONY: lint
lint:
	black .
	isort .
	ruff check .
	mypy xyphos_client

.PHONY: test
test:
	pytest --cov=xyphos_client

.PHONY: docs
docs:
	pdoc3 --html --output-dir docs xyphos_client --force

.PHONY: clean
clean:
	rm -rf build/
	rm -rf dist/
	rm -rf *.egg-info/
	rm -rf docs/
	rm -rf .pytest_cache/
	rm -rf .coverage
	rm -rf htmlcov/
	find . -type d -name "__pycache__" -exec rm -rf {} +
	find . -type d -name "*.pyc" -exec rm -rf {} +

.PHONY: build
build: clean
	python -m build

.PHONY: publish
publish: build
	twine upload dist/*

.PHONY: publish-test
publish-test: build
	twine upload --repository-url https://test.pypi.org/legacy/ dist/* 