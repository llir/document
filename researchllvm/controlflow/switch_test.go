package controlflow

import (
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"

	"github.com/dannypsnl/researchllvm/helper"
)

func TestSwitch(t *testing.T) {
	m := ir.NewModule()
	main := m.NewFunc("foo", types.Void)
	b := main.NewBlock("")

	compileStmt(b, &SSwitch{
		Target: &EBool{V: true},
		CaseList: []struct {
			Expr
			Stmt
		}{
			{Expr: &EBool{V: true}, Stmt: &SRet{Val: &EVoid{}}},
		},
		DefaultCase: &SRet{Val: &EVoid{}},
	})

	helper.PrettyPrint(m)
}
