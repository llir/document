.PHONY: deps serve
deps:
	pip3 install mkdocs --user
	nix-env -i graphviz
serve:
	@mkdocs serve

.PHONY: build commit publish
site:
	@git clone --branch gh-pages git@github.com:llir/document.git site
build: site
	mkdocs build

commit: build
	@cd site; git add -A
	@cd site; git commit -m "update $$(date +%Y/%m/%d-%H:%M:%S)"

publish: commit
	@cd site; git push
