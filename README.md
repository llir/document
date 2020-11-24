# LLVM Document

### Scripts

```shell script
# dependencies
make deps
# local test
make serve
# deployment
make publish
```

### Update Graph(If need)

```shell script
cd docs
dot -Tjpg ./classic.dot
dot -Tjpg ./llvm.dot
```