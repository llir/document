package controlflow

import (
	"github.com/dannypsnl/extend"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

type Expr interface{ isExpr() Expr }
type EVoid struct{ Expr }
type EBool struct {
	Expr
	V bool
}

type Stmt interface{ isStmt() Stmt }
type SBreak struct{ Stmt }
type SIf struct {
	Stmt
	Cond Expr
	Then Stmt
	Else Stmt
}
type SSwitch struct {
	Stmt
	Target   Expr
	CaseList []struct {
		Expr
		Stmt
	}
	DefaultCase Stmt
}
type SRet struct {
	Stmt
	Val Expr
}

func compileExpr(e Expr) value.Value {
	switch e := e.(type) {
	case *EBool:
		if e.V {
			return constant.NewInt(types.I1, 1)
		} else {
			return constant.NewInt(types.I1, 0)
		}
	case *EVoid:
		return nil
	}
	panic("unknown expression")
}

func compileStmt(block *ir.Block, stmt Stmt) {
	b := extend.Block(block)
	if !b.BelongsToFunc() {
		return
	}
	switch s := stmt.(type) {
	case *SIf:
		f := b.Parent
		thenB := extend.Block(f.NewBlock(""))
		compileStmt(thenB.Block, s.Then)
		elseB := f.NewBlock("")
		compileStmt(elseB, s.Else)
		b.NewCondBr(compileExpr(s.Cond), thenB.Block, elseB)
		if thenB.HasTerminator() {
			leaveB := f.NewBlock("")
			thenB.NewBr(leaveB)
		}
	case *SSwitch:
		addr := constant.NewBlockAddress(b.Parent, b)
		b.NewIndirectBr(addr)
	case *SRet:
		b.NewRet(compileExpr(s.Val))
	}
}
