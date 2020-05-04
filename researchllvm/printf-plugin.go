package researchllvm

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
)

func PrintfPlugin(mod *ir.Module) *ir.Func {
	printf := mod.NewFunc("printf", types.I32, ir.NewParam("format", types.NewPointer(types.I8)))
	printf.Sig.Variadic = true
	return printf
}
