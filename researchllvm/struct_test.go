package researchllvm

import (
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

func TestStruct(t *testing.T) {
	mod := ir.NewModule()

	stringTyp := mod.NewTypeDef(
		"string",
		types.NewStruct(
			types.NewPointer(types.I8),
		),
	)

	puts := mod.NewFunc(
		"puts",
		types.I32,
		ir.NewParam("format", types.NewPointer(types.I8)),
	)
	puts.Sig.Variadic = true

	helloWorldString := mod.NewGlobalDef("tmp", constant.NewCharArrayFromString("Hello, World!\n"))
	main := mod.NewFunc(
		"main",
		types.I32,
	)
	mainB := main.NewBlock("")
	ptrToStr := mainB.NewGetElementPtr(
		types.NewArray(14, types.I8), helloWorldString,
		constant.NewInt(types.I32, 0),
		constant.NewInt(types.I32, 0),
	)
	s := mainB.NewAlloca(stringTyp)
	sFieldCstring := mainB.NewGetElementPtr(
		stringTyp, s,
		constant.NewInt(types.I32, 0),
		constant.NewInt(types.I32, 0),
	)
	mainB.NewStore(ptrToStr, sFieldCstring)
	mainB.NewCall(puts, mainB.NewLoad(types.NewPointer(types.I8), sFieldCstring))
	mainB.NewRet(constant.NewInt(types.I32, 0))

	PrettyPrint(mod)

	executeIR(mod)
}

// generated LLVM IR:
//
// ```
// %string = type { i8* }
//
// @tmp = global [14 x i8] c"Hello, World!\0A"
//
// declare i32 @puts(i8* %format, ...)
//
// define i32 @main() {
// ; <label>:0
// 	%1 = getelementptr [14 x i8], [14 x i8]* @tmp, i32 0, i32 0
// 	%2 = alloca %string
// 	%3 = getelementptr %string, %string* %2, i32 0, i32 0
// 	store i8* %1, i8** %3
// 	%4 = load i8*, i8** %3
// 	%5 = call i32 (i8*, ...) @puts(i8* %4)
// 	ret i32 0
// }
// ```
// Output:
//
// Hello, World!
