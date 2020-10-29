.PHONY: build commit publish
build:
	mkdocs build

commit: build
	@cd site; git add -A
	@cd site; git commit -m "update $$(date +%Y/%m/%d-%H:%M:%S)"

publish: commit
	@cd site; git push
