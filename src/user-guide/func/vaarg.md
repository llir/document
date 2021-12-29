# Variant Argument (a.k.a. VAArg)

One example of a variadic function is `printf`. This is how to create a function prototype for `printf`:

```go
m := ir.NewModule()

printf := m.NewFunc(
	"printf",
	types.I32,
	ir.NewParam("", types.NewPointer(types.I8)),
)
printf.Sig.Variadic = true
```

The above code would produce the following IR:

```llvm
declare i32 @printf(i8*, ...)
```
