package controlflow

import (
	"fmt"
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
)

func TestDoWhile(t *testing.T) {
	f := ir.NewFunc("foo", types.Void)
	b := f.NewBlock("")

	compileStmt(b, &SDoWhile{
		Cond: &EBool{V: true},
		Block: &SDefine{
			Stmt: nil,
			Name: "foo",
			Typ:  types.I32,
			Expr: &EI32{V: 1},
		},
	})

	f.Blocks[len(f.Blocks)-1].NewRet(nil)

	fmt.Println(f.LLString())
}
