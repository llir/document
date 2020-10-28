package controlflow

import (
	"fmt"
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
)

func TestForLoop(t *testing.T) {
	f := ir.NewFunc("foo", types.Void)
	ctx := NewContext(f.NewBlock(""))

	ctx.compileStmt(&SForLoop{
		InitName: "x",
		InitExpr: &EI32{V: 0},
		Step:     &EAdd{Lhs: &EVariable{Name: "x"}, Rhs: &EI32{V: 1}},
		Cond:     &ELessThan{Lhs: &EVariable{Name: "x"}, Rhs: &EI32{V: 10}},
		Block:    &SDefine{Name: "foo", Typ: types.I32, Expr: &EI32{V: 2}},
	})

	f.Blocks[len(f.Blocks)-1].NewRet(nil)

	fmt.Println(f.LLString())
}
