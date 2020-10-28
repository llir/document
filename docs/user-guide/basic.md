# Basic Introduction

## Module

An LLVM IR file is a module. A module has many top-level entities:

- global variables
- functions
- types
- metadata

In this basic introduction, we won't dig into metadata, but instead focus on what we can do with global variables, functions, and types.

[llir/llvm](https://github.com/llir/llvm) provides package [ir](https://pkg.go.dev/github.com/llir/llvm/ir?tab=doc) for these concepts. Let's see how a C program can be translated into LLVM IR using [llir/llvm](https://github.com/llir/llvm).

C example:

```c
int g = 2;

int add(int x, int y) {
  return x + y;
}
int main() {
  return add(1, g);
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
	mb.NewRet(mb.NewCall(funcAdd, constant.NewInt(types.I32, 1), mb.NewLoad(types.I32, globalG)))

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
	ret i32 %2
}
```

In this example, we have one global variable and two functions, mapping to C code. Now let's dig into global variables.

## Global Variable

In LLVM IR assembly, the identifier of global variables are prefixed with an `@` character.
Importantly, global variables are represented in LLVM as pointers, so we have to use [load](https://pkg.go.dev/github.com/llir/llvm/ir?tab=doc#InstLoad) to retreive the value and [store](https://pkg.go.dev/github.com/llir/llvm/ir?tab=doc#InstStore) to update the value of a global variable.

## Function

Like globals, in LLVM IR assembly the identifier of functions are prefixed with an `@` character. Functions composed by a function prototype and a group of basic blocks.
A function without basic blocks is a function declaration. The following code would generate a function declaration:

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
declare i32 @add(i32 %0, i32 %1)
```

When we want to bind to existing functions defined in other object files, we would create function declarations.

### Function Prototype

A function prototype or function signature defines the parameters and return type of a function.

### Basic Block

If function is group of basic blocks, then a basic block is a group of instructions. The basic notion behind a basic block is that if any instruction of a basic block is executed, then all instructions of the basic block are executed. In other words, there may be no branching or terminating instruction in the middle of a basic block, and all incoming branches must transfer control flow to the first instruction of the basic block.

It is worthwhile to note that most high-level expression would be lowered into a set of instructions, covering one or more basic blocks.

[llir/llvm](https://github.com/llir/llvm) provides API to create instructions by a basic block.
For further information, refer to the [Block API documentation](https://pkg.go.dev/github.com/llir/llvm/ir?tab=doc#Block).

### Instruction

An instruction is a set of operations on assembly abstraction level which operate on an abstract machine model, as defined by LLVM.
For further information, refer to the [Instruction Reference section of the LLVM Language Reference Manual](https://llvm.org/docs/LangRef.html#instruction-reference).

## Type

There are many types in LLVM type system, here we focus on how to create a new type.

```go
m := ir.NewModule()

m.NewTypeDef("foo", types.NewStruct(types.I32))
```

The above code would produce the following IR:

```llvm
%foo = type { i32 }
```

Which could be mapped to the following C code:

```c
typedef struct {
  int x;
} foo;
```

Notice that in LLVM, structure fields have no name.

## Conclusion

We hope that the previous sections have provide enough information about how to get use the documentation to dig into details.
We will not dig into the details of each instruction; instead, we aim to provide a whole picture about how to use the library.
Therefore, the next section is a list of common high-level concept and how to map them to IR.
