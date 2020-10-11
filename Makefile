.PHONY: build
build:
	mkdocs build

.PHONY: commit
commit: build
	@cd site; git add -A
	@cd site; git commit -m "update $$(date +%Y/%m/%d-%H:%M:%S)"
