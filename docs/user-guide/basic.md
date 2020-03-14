# Basic Introduction

## Module

A LLVM IR file is a module. A module owns many global level components:

- global variable
- function
- type
- metadata

In this basic introduction, we don't dig into metadata, but focus on what can we do with global variable, function, and type.

[llir/llvm](https://github.com/llir/llvm) provides package `ir` for these concepts, let's see what can a C program being translated to LLVM IR using [llir/llvm](https://github.com/llir/llvm).

C example:

```c
int g = 2;

int add(int x, int y) {
  return x + y;
}
int main() {
  add(1, g);
  return 0;
}
```

Generate module:

```go
package main

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

func main() {
	m := ir.NewModule()

	globalG := m.NewGlobalDef("g", constant.NewInt(types.I32, 2))

	funcAdd := m.NewFunc("add", types.I32,
		ir.NewParam("x", types.I32),
		ir.NewParam("y", types.I32),
	)
	ab := funcAdd.NewBlock("")
	ab.NewRet(ab.NewAdd(funcAdd.Params[0], funcAdd.Params[1]))

	funcMain := m.NewFunc(
		"main",
		types.I32,
	)  // omit parameters
	mb := funcMain.NewBlock("") // llir/llvm would give correct default name for block without name
	mb.NewCall(funcAdd, constant.NewInt(types.I32, 1), mb.NewLoad(types.I32, globalG))
	mb.NewRet(constant.NewInt(types.I32, 0))

	println(m.String())
}
```

Generated IR:

```llvm
@g = global i32 2

define i32 @add(i32 %x, i32 %y) {
; <label>:0
	%1 = add i32 %x, %y
	ret i32 %1
}

define i32 @main() {
; <label>:0
	%1 = load i32, i32* @g
	%2 = call i32 @add(i32 1, i32 %1)
	ret i32 0
}
```

In this example, we have global variable and function, mapping to C code. Now we dig into global variable.

## Global Variable

Globals prefixed with `@` character.
An important thing is globals in LLVM, is a pointer, so have to `load` for its value,`store` to update its value.

## Function

As globals, function name prefixed with `@` character. Function composed by prototype and a group of basic blocks.
If there has no basic block, then a function is a declaration, the following code would generate a declaration:

```go
m.NewFunc(
    "add",
	types.I32,
    ir.NewParam("", types.I32),
	ir.NewParam("", types.I32),
)
```

Output:

```llvm
declare i32 @add(i32, i32)
```

When we want to bind to existed function in others object files, we would create a declaration.

### Prototype

Prototype means parameters and return type.

### Basic Block

If function is group of basic blocks, then basic blocks is a group of instructions.
An important thing is most high-level expression would break down into few instructions.

[llir/llvm](https://github.com/llir/llvm) provides API to create instructions by a basic block.
To get more information, goto [Block API document](https://pkg.go.dev/github.com/llir/llvm@v0.3.0/ir?tab=doc#Block).

### Instruction

Instruction is a set of operations on assembly abstraction level to operate on an abstract machine model.
To get more information, goto [LLVM Language Reference Manual: instruction reference](https://llvm.org/docs/LangRef.html#instruction-reference).

## Type

There are many types in LLVM type system, here focus on how to create a new type.

```go
m := ir.NewModule()

m.NewTypeDef("foo", types.NewStruct(types.I32))
```

Above code would produce:

```llvm
%foo = type { i32 }
```

It could map to C code:

```c
struct foo {
  int x;
};
```

Notice in LLVM, structure field has no name.

## Conclusion

Hope previous sections provide enough information about how to get enough information to dig into details.
We will not dig into the details of each instruction, instead of that, we would provide a whole picture about how to use the library.
Therefore, the next section is a list of common high-level concept and how to map them to IR.