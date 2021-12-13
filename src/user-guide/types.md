# High level types

## Structure

Structure is quite common and basic type in programming language. Here focus on how to create an equal LLVM mapping for structure.

LLVM has the concept about structure type, but structure type in LLVM didn't have field name, how to get/set fields of structure would be a thing. Let's assume a structure: `foo` has a field named `x` with type `i32`. Below code shows how to access `x`.

```go
zero := constant.NewInt(types.I32, 0)
m := ir.NewModule()

foo := m.NewTypeDef("foo", types.NewStruct(types.I32))

main := m.NewFunc("main", types.I32)
b := main.NewBlock("")
fooInstance := b.NewAlloca(foo)
fieldX := b.NewGetElementPtr(foo, fooInstance, zero, zero)
// now `fieldX` is a pointer to field `x` of `foo`.
b.NewStore(constant.NewInt(types.I32, 10), fieldX)
b.NewLoad(types.I32, fieldX)
b.NewRet(zero)
```

Get element pointer(a.k.a. GEP) is the point here.
It computes a pointer to any structure member with no overhead.
Then `load` and `store` can work on it.

To get more information about GEP can goto [Lang Ref](https://llvm.org/docs/LangRef.html#getelementptr-instruction).

## Array

Arrays are a sequence of elements of the same types in LLVM. And in LLVM, array size is a constant that never changes. In the following text, I will present how to define a global array, and how to operate it.

As usual, we need a module and some helping code

```go
mod := ir.NewModule()
printf := PrintfPlugin(mod)
formatString := "array_def[%d]: %d\n"
fmtStr := mod.NewGlobalDef("x", constant.NewCharArrayFromString(formatString))
main := mod.NewFunc("main", types.I32)
mainB := main.NewBlock("")
ptrToStr := mainB.NewGetElementPtr(types.NewArray(uint64(len(formatString)), types.I8), fmtStr, CI32(0), CI32(0))
```

Now we create a global array

```go
arrTy := types.NewArray(5, types.I8)
arrayDef := mod.NewGlobalDef("array_def", constant.NewArray(arrTy, CI8(1), CI8(2), CI8(3), CI8(4), CI8(5)))
```

Then we can load it into stack, in the following code, we

1. extract value from loaded array
2. insert value into loaded array

```go
arr := mainB.NewLoad(arrTy, arrayDef)
for i := 0; i < 5; i++ {
	mainB.NewCall(printf, ptrToStr, CI32(int64(i)), mainB.NewExtractValue(arr, uint64(i)))
	mainB.NewInsertValue(arr, CI8(0), uint64(i))
	mainB.NewCall(printf, ptrToStr, CI32(int64(i)), mainB.NewExtractValue(arr, uint64(i)))
}
```

Another way to get element and update it is using get element pointer(a.k.a GEP)

```go
for i := 0; i < 5; i++ {
	pToElem := mainB.NewGetElementPtr(arrTy, arrayDef, CI32(0), CI32(int64(i)))
	mainB.NewCall(printf, ptrToStr, CI32(int64(i)), mainB.NewLoad(types.I8, pToElem))
	mainB.NewStore(CI8(0), pToElem)
	mainB.NewCall(printf, ptrToStr, CI32(int64(i)), mainB.NewLoad(types.I8, pToElem))
}
```

compare with previous result, you will find previous insert didn't work as expected. So this is the key point about insert value. It don't update your aggregate value, it do thing like immutable programming: take and return new one. Thus, we should write

```go
for i := 0; i < 5; i++ {
	mainB.NewCall(printf, ptrToStr, CI32(int64(i)), mainB.NewExtractValue(arr, uint64(i)))
	newArr := mainB.NewInsertValue(arr, CI8(0), uint64(i))
	mainB.NewCall(printf, ptrToStr, CI32(int64(i)), mainB.NewExtractValue(newArr, uint64(i)))
}
```

Now you should understand array enough to compile your language into this aggregate type!

## Algebra Data Type

Algebra data type is a common concept in functional programming language.
For example, Haskell can write:

```hs
data Expr =
  EInt Int
  | EBool Bool
  | EString String
```

How to express such concept in LLVM?
The idea was selecting the biggest size in all variants and use it as the size of this type.
Do `bitcast` when need to access the actual value.
Here is the code:

```go
mod := ir.NewModule()

typeExpr := mod.NewTypeDef("Expr", types.NewStruct(
	types.I8,
	types.NewArray(8, types.I8),
))
// variant tag 0
typeExprInt := mod.NewTypeDef("EInt", types.NewStruct(
	types.I8,
	types.I32,
))
// variant tag 1
mod.NewTypeDef("EBool", types.NewStruct(
	types.I8,
	types.I1,
))
// variant tag 2
mod.NewTypeDef("EString", types.NewStruct(
	types.I8,
	types.NewPointer(types.I8),
))

main := mod.NewFunc(
	"main",
	types.I32,
)
b := main.NewBlock("")
exprInstance := b.NewAlloca(typeExpr)
exprTag := b.NewGetElementPtr(typeExpr, exprInstance, constI32(0), constI32(0))
// set tag to 0
b.NewStore(constI8(0), exprTag)
exprIntInstance := b.NewBitCast(exprInstance, types.NewPointer(typeExprInt))
exprIntValue := b.NewGetElementPtr(typeExprInt, exprIntInstance, constI32(0), constI32(1))
b.NewStore(constI32(2), exprIntValue)
b.NewRet(constI32(0))
```

These code produce:

```llvm
%Expr = type { i8, [8 x i8] }
%EInt = type { i8, i32 }
%EBool = type { i8, i1 }
%EString = type { i8, i8* }

define i32 @main() {
; <label>:0
	%1 = alloca %Expr
	%2 = getelementptr %Expr, %Expr* %1, i32 0, i32 0
	store i8 0, i8* %2
	%3 = bitcast %Expr* %1 to %EInt*
	%4 = getelementptr %EInt, %EInt* %3, i32 0, i32 1
	store i32 2, i32* %4
	ret i32 0
}
```

`tag` in each variant is important, for example, pattern matching in Haskell looks like:

```hs
case expr of
  EInt i -> "int"
  EBool b -> "bool"
  EString s -> "string"
```

`case` expression would need tag to distinguish variant.
