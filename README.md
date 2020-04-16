# LLVM Document

### Scripts

```shell script
# dependencies
pip3 install mkdocs
pip3 install markdown_inline_graphviz_extension --user
# local test
mkdocs serve
# Clone repo again to `site/`
git clone git@github.com:llir/document.git site
# checkout to deploy branch
cd site && git checkout gh-pages
cd ..
# deployment
mkdocs build
```
