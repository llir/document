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
type EVoid struct{ Expr }
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
	Then Stmt
	Else Stmt
}
type SRet struct {
	Stmt
	Val Expr
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
		return b.NewICmp(enum.IPredEQ, compileExpr(b, e.L), compileExpr(b, e.R))
	case *EVoid:
		return nil
	}
	panic("unknown expression")
}

func compileStmt(f *ir.Func, bb *ir.Block, stmt Stmt) {
	switch s := stmt.(type) {
	case *SIf:
		thenB := f.NewBlock("")
		compileStmt(f, thenB, s.Then)
		elseB := f.NewBlock("")
		compileStmt(f, elseB, s.Else)
		bb.NewCondBr(compileExpr(bb, s.Cond), thenB, elseB)
		if thenB.Term == nil {
			leaveB := f.NewBlock("")
			thenB.NewBr(leaveB)
		}
	case *SRet:
		bb.NewRet(compileExpr(bb, s.Val))
	}
}

func TestParameterAttr(t *testing.T) {
	f := ir.NewFunc("foo", types.Void)
	bb := f.NewBlock("")

	compileStmt(f, bb, &SIf{
		Cond: &EBool{V: true},
		Then: nil,
		Else: &SRet{Val: &EVoid{}},
	})

	// whatever what we did in compileStmt, we use convention that a block leave in the end is empty.
	f.Blocks[len(f.Blocks)-1].NewRet(nil)

	fmt.Println(f.LLString())
}

// generated LLVM IR:
//
// ```
// define void @foo() {
// ; <label>:0
// br i1 true, label %1, label %2
//
// ; <label>:1
// br label %3
//
// ; <label>:2
// ret void
//
// ; <label>:3
// ret void
// }
// ```
