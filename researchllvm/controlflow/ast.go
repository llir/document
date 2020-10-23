package controlflow

import (
	"github.com/dannypsnl/extend"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

type Expr interface{ isExpr() Expr }
type EVoid struct{ Expr }
type EBool struct {
	Expr
	V bool
}
type EI32 struct {
	Expr
	V int64
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
type SDoWhile struct {
	Stmt
	Cond  Expr
	Block Stmt
}
type SDefine struct {
	Stmt
	Name string
	Typ  types.Type
	Expr Expr
}
type SRet struct {
	Stmt
	Val Expr
}

func compileConstant(e Expr) constant.Constant {
	switch e := e.(type) {
	case *EI32:
		return constant.NewInt(types.I32, e.V)

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
	f := b.Parent
	switch s := stmt.(type) {
	case *SIf:
		thenB := extend.Block(f.NewBlock(""))
		compileStmt(thenB.Block, s.Then)
		elseB := f.NewBlock("")
		compileStmt(elseB, s.Else)
		b.NewCondBr(compileConstant(s.Cond), thenB.Block, elseB)
		if thenB.HasTerminator() {
			leaveB := f.NewBlock("")
			thenB.NewBr(leaveB)
		}
	case *SSwitch:
		cases := []*ir.Case{}
		for _, ca := range s.CaseList {
			caseB := f.NewBlock("")
			compileStmt(caseB, ca.Stmt)
			cases = append(cases, ir.NewCase(compileConstant(ca.Expr), caseB))
		}
		defaultB := f.NewBlock("")
		compileStmt(defaultB, s.DefaultCase)
		b.NewSwitch(compileConstant(s.Target), defaultB, cases...)
	case *SDoWhile:
		doB := b.Block
		// if previous block is not empty, then we need to create new block for do-while loop
		if b.Insts != nil {
			doB = f.NewBlock("")
		}
		compileStmt(doB, s.Block)
		leaveB := f.NewBlock("")
		doB.NewCondBr(compileConstant(s.Cond), doB, leaveB)
	case *SDefine:
		v := b.NewAlloca(s.Typ)
		v.SetName(s.Name)
		b.NewStore(compileConstant(s.Expr), v)
	case *SRet:
		b.NewRet(compileConstant(s.Val))
	}
}
