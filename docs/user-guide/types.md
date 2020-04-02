# High level types

## Structure

Structure is quite common and basic type in programming language.
Here focus on how to create an equal LLVM mapping for structure.

LLVM has the concept about structure type, but structure type in LLVM didn't have field name, how to get/set fields of structure would be a thing.
Let's assume a structure: `foo` has a field named `x` with type `i32`.
Below code shows how to access `x`.

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
