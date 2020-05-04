package researchllvm

import (
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

func TestMalloc(t *testing.T) {
	mod := ir.NewModule()

	structType := mod.NewTypeDef(
		"foo",
		types.NewStruct(
			types.NewPointer(types.I8),
			types.I64,
		),
	)

	mallocFunc := mod.NewFunc("malloc",
		types.NewPointer(types.I8),
		ir.NewParam("", types.I64),
	)

	main := mod.NewFunc(
		"main",
		types.I32,
	)
	block := main.NewBlock("")
	mallocatedSpaceRaw := block.NewCall(mallocFunc, constant.NewInt(types.I64, 128))
	block.NewBitCast(mallocatedSpaceRaw, types.NewPointer(structType))
	block.NewRet(constant.NewInt(types.I32, 0))

	PrettyPrint(mod)
}

// generated LLVM IR:
//
// ```
// %foo = type { i8*, i64 }
//
// declare i8* @malloc(i64)
//
// define i32 @main() {
// ; <label>:0
// 	%1 = call i8* @malloc(i64 128)
// 	%2 = bitcast i8* %1 to %foo*
// 	ret i32 0
// }
// ```
