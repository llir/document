package controlflow

import (
	"fmt"
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
)

func TestDoWhile(t *testing.T) {
	f := ir.NewFunc("foo", types.Void)
	ctx := NewContext(f.NewBlock(""))

	ctx.compileStmt(&SDoWhile{
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
