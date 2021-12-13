package researchllvm

import (
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/researchllvm/helper"
)

func TestVAArg(t *testing.T) {
	m := ir.NewModule()

	printf := m.NewFunc(
		"printf",
		types.I32,
		ir.NewParam("", types.NewPointer(types.I8)),
	)
	printf.Sig.Variadic = true

	helper.PrettyPrint(m)
}
