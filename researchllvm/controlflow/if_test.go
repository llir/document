package controlflow

import (
	"fmt"
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

type Expr interface{ isExpr() Expr }
type EBool struct {
	Expr
	V bool
}
type EEq struct {
	// `==`
	Expr
	L, R Expr
}

type Stmt interface{ isStmt() Stmt }
type SIf struct {
	Stmt
	Cond Expr
}
type SAssign struct {
	Stmt
	VarName string
	Val     Expr
}

func compileExpr(b *ir.Block, e Expr) value.Value {
	switch e := e.(type) {
	case *EBool:
		if e.V {
			return constant.NewInt(types.I1, 1)
		} else {
			return constant.NewInt(types.I1, 0)
		}
	case *EEq:
		b.NewICmp(enum.IPredEQ, compileExpr(b, e.L), compileExpr(b, e.R))
	}
	panic("unknown expression")
}

func compileIf(f *ir.Func, ifStmt *SIf) {
	b := f.NewBlock("if")
	cond := compileExpr(b, ifStmt.Cond)
	tru := f.NewBlock("if-then")
	fls := f.NewBlock("if-else")
	b.NewCondBr(cond, tru, fls)
	tru.NewRet(nil)
	fls.NewRet(nil)
}

func TestParameterAttr(t *testing.T) {
	f := ir.NewFunc("foo", types.Void)

	compileIf(f, &SIf{
		Cond: &EBool{V: true},
	})

	fmt.Println(f.LLString())
}

// generated LLVM IR:
//
// ```
// %Foo = type { i32 }
//
// declare void @foo(%Foo noalias sret %result)
// ```
