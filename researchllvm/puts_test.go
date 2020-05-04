package researchllvm

import (
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

func TestPuts(t *testing.T) {
	mod := ir.NewModule()

	helloWorldString := mod.NewGlobalDef("tmp", constant.NewCharArrayFromString("Hello, World!\n"))

	puts := mod.NewFunc(
		"puts",
		types.I32,
		ir.NewParam("format", types.NewPointer(types.I8)),
	)
	puts.Sig.Variadic = true

	main := mod.NewFunc(
		"main",
		types.I32,
	)
	mainB := main.NewBlock("")
	pointerToString := mainB.NewGetElementPtr(
		types.NewArray(14, types.I8), helloWorldString,
		constant.NewInt(types.I32, 0),
		constant.NewInt(types.I32, 0),
	)
	mainB.NewCall(puts, pointerToString)
	mainB.NewRet(constant.NewInt(types.I32, 0))

	PrettyPrint(mod)

	executeIR(mod)
}

// generated LLVM IR:
//
// ```
// @tmp = global [14 x i8] c"Hello, World!\0A"
//
// declare i32 @puts(i8* %format, ...)
//
// define i32 @main() {
// ; <label>:0
// 	%1 = getelementptr [14 x i8], [14 x i8]* @tmp, i32 0, i32 0
// 	%2 = call i32 (i8*, ...) @puts(i8* %1)
// 	ret i32 0
// }
// ```
// Output:
//
// Hello, World!
