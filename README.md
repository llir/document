# LLVM Document

### Scripts

```shell script
# dependencies
pip3 install mkdocs
pip3 install markdown_inline_graphviz_extension --user
# NOTE: Nix works on all Unix-like machine, therefore, I pick its command as template
# You can pick any package manager you familiar with. Just remember install `graphviz`
nix-env -i graphviz
# local test
mkdocs serve
# Clone repo again to `site/`
git clone --branch gh-pages git@github.com:llir/document.git site
# deployment
mkdocs build
```
