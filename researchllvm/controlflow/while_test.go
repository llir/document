package controlflow

import (
	"fmt"
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
)

func TestWhile(t *testing.T) {
	f := ir.NewFunc("foo", types.Void)
	ctx := NewContext(f.NewBlock(""))

	ctx.compileStmt(&SWhile{
		Cond: &EBool{V: true},
		Block: &SDefine{
			Name: "x",
			Typ:  types.I32,
			Expr: &EI32{V: 0},
		},
	})

	f.Blocks[len(f.Blocks)-1].NewRet(nil)

	fmt.Println(f.LLString())
}
