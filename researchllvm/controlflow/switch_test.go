package controlflow

import (
	"fmt"
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
)

func TestSwitch(t *testing.T) {
	f := ir.NewFunc("foo", types.Void)
	b := f.NewBlock("")

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

	fmt.Println(f.LLString())
}
