package researchllvm

import (
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
)

func TestParameterAttr(t *testing.T) {
	m := ir.NewModule()

	fooTyp := m.NewTypeDef("Foo", types.NewStruct(
		types.I32,
	))
	retS := ir.NewParam("result", fooTyp)
	retS.Attrs = append(retS.Attrs, enum.ParamAttrNoAlias)
	retS.Attrs = append(retS.Attrs, enum.ParamAttrSRet)
	m.NewFunc("foo", types.Void, retS)

	PrettyPrint(m)
}

// generated LLVM IR:
//
// ```
// %Foo = type { i32 }
//
// declare void @foo(%Foo noalias sret %result)
// ```
