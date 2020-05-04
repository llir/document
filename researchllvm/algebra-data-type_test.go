package researchllvm

import (
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

func TestAlgebraDataType(t *testing.T) {
	mod := ir.NewModule()

	typeExpr := mod.NewTypeDef("Expr", types.NewStruct(
		types.I8,
		types.NewArray(8, types.I8),
	))
	typeExprInt := mod.NewTypeDef("EInt", types.NewStruct(
		types.I8,
		types.I32,
	))
	mod.NewTypeDef("EBool", types.NewStruct(
		types.I8,
		types.I1,
	))
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

	PrettyPrint(mod)
}

func constI8(v int64) constant.Constant {
	return constant.NewInt(types.I8, v)
}
func constI32(v int64) constant.Constant {
	return constant.NewInt(types.I32, v)
}
