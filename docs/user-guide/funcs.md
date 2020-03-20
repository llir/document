# More Function

### Linkage

The following code shows some linkage can use in IR.

```go
m := ir.NewModule()

add := m.NewFunc("add", types.I64, ir.NewParam("", types.I64))
add.Linkage = enum.LinkageInternal
add = m.NewFunc("add1", types.I64, ir.NewParam("", types.I64))
add.Linkage = enum.LinkageLinkOnce
add = m.NewFunc("add2", types.I64, ir.NewParam("", types.I64))
add.Linkage = enum.LinkagePrivate
add = m.NewFunc("add3", types.I64, ir.NewParam("", types.I64))
add.Linkage = enum.LinkageWeak
add = m.NewFunc("add4", types.I64, ir.NewParam("", types.I64))
add.Linkage = enum.LinkageExternal
```

The code would produce:

```llvm
declare internal i64 @add(i64)

declare linkonce i64 @add1(i64)

declare private i64 @add2(i64)

declare weak i64 @add3(i64)

declare external i64 @add4(i64)
```

To get more information about linkage, read [llvm doc](https://llvm.org/docs/LangRef.html#linkage-types) and [pkg.go.dev](https://pkg.go.dev/github.com/llir/llvm/ir/enum?tab=doc#Linkage).

### Variant Argument(a.k.a. VAArg)

One example is `printf`:

```go
m := ir.NewModule()

printf := m.NewFunc(
	"printf",
	types.I32,
	ir.NewParam("", types.NewPointer(types.I8)),
)
printf.Sig.Variadic = true
```

The code would produce:

```llvm
declare i32 @printf(i8*, ...)
```

### Function Overloading

There has no overloading in IR, therefore solution is creating two functions.