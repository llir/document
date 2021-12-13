# More Function

## Linkage

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

## Variant Argument (a.k.a. VAArg)

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

## Function Overloading

There is no overloading in LLVM IR. One solution is to create one function per function signature, where each LLVM IR function would have a unique name (this is why C++ compilers do name mangling).

## First-class Function(Closure)

### Naive Implementation

Create a closure(a.k.a. first-class function) requires a place to store captured variables. In LLVM, the best way is create a structure for such case:

```go
m := ir.NewModule()

zero := constant.NewInt(types.I32, 0)
one := constant.NewInt(types.I32, 1)

captureStruct := m.NewTypeDef("id_capture", types.NewStruct(
	types.I32,
))
captureTyp := types.NewPointer(captureStruct)
idFn := m.NewFunc("id", types.I32, ir.NewParam("capture", captureTyp))
idB := idFn.NewBlock("")
v := idB.NewGetElementPtr(captureStruct, idFn.Params[0], zero, zero)
idB.NewRet(idB.NewLoad(types.I32, v))
idClosureTyp := m.NewTypeDef("id_closure", types.NewStruct(
	captureTyp,
	idFn.Type(),
))

mainFn := m.NewFunc("main", types.I32)
b := mainFn.NewBlock("")
// define a local variable `i`
i := b.NewAlloca(types.I32)
b.NewStore(constant.NewInt(types.I32, 10), i)
// use alloca at here to simplify code, in real case should be `malloc` or `gc_malloc`
captureInstance := b.NewAlloca(captureStruct)
ptrToCapture := b.NewGetElementPtr(captureStruct, captureInstance, zero, zero)
// capture variable
b.NewStore(b.NewLoad(types.I32, i), ptrToCapture)
// prepare closure
idClosure := b.NewAlloca(idClosureTyp)
ptrToCapturePtr := b.NewGetElementPtr(idClosureTyp, idClosure, zero, zero)
b.NewStore(captureInstance, ptrToCapturePtr)
ptrToFuncPtr := b.NewGetElementPtr(idClosureTyp, idClosure, zero, one)
b.NewStore(idFn, ptrToFuncPtr)
// assuming we transfer closure into another context
accessCapture := b.NewGetElementPtr(idClosureTyp, idClosure, zero, zero)
accessFunc := b.NewGetElementPtr(idClosureTyp, idClosure, zero, one)
result := b.NewCall(b.NewLoad(idFn.Type(), accessFunc), b.NewLoad(captureTyp, accessCapture))

printIntegerFormat := m.NewGlobalDef("tmp", constant.NewCharArrayFromString("%d\n"))
pointerToString := b.NewGetElementPtr(types.NewArray(3, types.I8), printIntegerFormat, zero, zero)
// ignore printf
b.NewCall(printf, pointerToString, result)

b.NewRet(constant.NewInt(types.I32, 0))
```

This is a huge example, I understand it's hard to read, but concept is clean. It would generate below LLVM IR:

```llvm
%id_capture = type { i32 }
%id_closure = type { %id_capture*, i32 (%id_capture*)* }

@tmp = global [3 x i8] c"%d\0A"

declare i32 @printf(i8* %format, ...)

define i32 @id(%id_capture* %capture) {
; <label>:0
	%1 = getelementptr %id_capture, %id_capture* %capture, i32 0, i32 0
	%2 = load i32, i32* %1
	ret i32 %2
}

define i32 @main() {
; <label>:0
	%1 = alloca i32
	store i32 10, i32* %1
	%2 = alloca %id_capture
	%3 = getelementptr %id_capture, %id_capture* %2, i32 0, i32 0
	%4 = load i32, i32* %1
	store i32 %4, i32* %3
	%5 = alloca %id_closure
	%6 = getelementptr %id_closure, %id_closure* %5, i32 0, i32 0
	store %id_capture* %2, %id_capture** %6
	%7 = getelementptr %id_closure, %id_closure* %5, i32 0, i32 1
	store i32 (%id_capture*)* @id, i32 (%id_capture*)** %7
	%8 = getelementptr %id_closure, %id_closure* %5, i32 0, i32 0
	%9 = getelementptr %id_closure, %id_closure* %5, i32 0, i32 1
	%10 = load i32 (%id_capture*)*, i32 (%id_capture*)** %9
	%11 = load %id_capture*, %id_capture** %8
	%12 = call i32 %10(%id_capture* %11)
	%13 = getelementptr [3 x i8], [3 x i8]* @tmp, i32 0, i32 0
	%14 = call i32 (i8*, ...) @printf(i8* %13, i32 %12)
	ret i32 0
}
```

Our `id` function captures an Integer and return it. To reach that `id_capture` was introduced for storing captured value. For passing whole closure in convenience, `id_closure` was introduced and stored capture structure and function pointer. When invoke a closure, get captured structure and function pointer from `id_closure` structure, then apply function with captured structure and additional arguments(if there's any). In this example omit the part about memory management, all structures allocated in the stack, this won't work in most real world case. Must notice this problem.

### Improvements

The naive implementation is not good enough, we have several ways can improve it, but instead of implementing them I'm going to list what can we do:

- Laziness function: Arity would be a thing in case
- Access cross asynchronous model
- If language has copy capture and reference capture, e.g. C++?
- What if working with a GC?

## Return Structure

When meet program that return structure by value, compiler has chance to remove such cloning. That's storing return structure into a reference passed by the caller. Which means, if we get:

```c
struct Foo {
    // ...
};

Foo foo() {
    Foo f;
    // ...
    return f;
}
```

should compile to:

```llvm
define void @foo(%Foo* noalias sret f) {
    // ...
}
```

- `sret` hints this is a return value.
- `noalias` hints other arguments won't point to the same place, LLVM optimizer might rely on such fact, so don't add it everywhere.

### Add parameter attributes

Here is example shows how to add parameter attributes:

```go
m := ir.NewModule()

fooTyp := m.NewTypeDef("Foo", types.NewStruct(
	types.I32,
))
retS := ir.NewParam("result", fooTyp)
retS.Attrs = append(retS.Attrs, enum.ParamAttrNoAlias)
retS.Attrs = append(retS.Attrs, enum.ParamAttrSRet)
m.NewFunc("foo", types.Void, retS)
```
