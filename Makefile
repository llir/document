.PHONY: deps serve
deps:
	pip3 install mkdocs --user
	nix-env -i graphviz
serve:
	@mkdocs serve
