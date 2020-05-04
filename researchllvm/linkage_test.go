package researchllvm

import (
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
)

func TestLinkage(t *testing.T) {
	m := ir.NewModule()

	add := m.NewFunc("add", types.I64, ir.NewParam("", types.I64))
	add.Linkage = enum.LinkageInternal
	add = m.NewFunc("add1", types.I64, ir.NewParam("", types.I64))
	add.Linkage = enum.LinkageLinkOnce
	add = m.NewFunc("add2", types.I64, ir.NewParam("", types.I64))
	add.Linkage = enum.LinkagePrivate
	add = m.NewFunc("add3", types.I64, ir.NewParam("", types.I64))
	add.Linkage = enum.LinkageWeak
	add = m.NewFunc("add4", types.I64, ir.NewParam("", types.I64))
	add.Linkage = enum.LinkageExternal

	PrettyPrint(m)
}
