# Linkage

The following code shows some linkage can use in IR.

```go
m := ir.NewModule()

add := m.NewFunc("add", types.I64, ir.NewParam("", types.I64))
add.Linkage = enum.LinkageInternal
add1 := m.NewFunc("add1", types.I64, ir.NewParam("", types.I64))
add1.Linkage = enum.LinkageLinkOnce
add2 := m.NewFunc("add2", types.I64, ir.NewParam("", types.I64))
add2.Linkage = enum.LinkagePrivate
add3 := m.NewFunc("add3", types.I64, ir.NewParam("", types.I64))
add3.Linkage = enum.LinkageWeak
add4 := m.NewFunc("add4", types.I64, ir.NewParam("", types.I64))
add4.Linkage = enum.LinkageExternal
```

The code would produce:

```llvm
declare internal i64 @add(i64)

declare linkonce i64 @add1(i64)

declare private i64 @add2(i64)

declare weak i64 @add3(i64)

declare external i64 @add4(i64)
```

For further information about linkage, refer to [LLVM doc](https://llvm.org/docs/LangRef.html#linkage-types) and [pkg.go.dev](https://pkg.go.dev/github.com/llir/llvm/ir/enum?tab=doc#Linkage).
