package controlflow

import (
	"fmt"

	"github.com/dannypsnl/extend"
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
	. "github.com/llir/researchllvm/helper"
)

type Expr interface{ isExpr() Expr }
type EConstant interface {
	Expr
	isEConstant() EConstant
}
type EVoid struct{ EConstant }
type EBool struct {
	EConstant
	V bool
}
type EI32 struct {
	EConstant
	V int64
}
type EVariable struct {
	Expr
	Name string
}
type EAdd struct {
	Expr
	Lhs, Rhs Expr
}
type ELessThan struct {
	Expr
	Lhs, Rhs Expr
}

func compileConstant(e EConstant) constant.Constant {
	switch e := e.(type) {
	case *EI32:
		return CI32(e.V)
	case *EBool:
		if e.V {
			return CI1(1)
		} else {
			return CI1(0)
		}
	case *EVoid:
		return nil
	}
	panic("unknown expression")
}

func (ctx *Context) compileExpr(e Expr) value.Value {
	switch e := e.(type) {
	case *EVariable:
		return ctx.lookupVariable(e.Name)
	case *EAdd:
		l, r := ctx.compileExpr(e.Lhs), ctx.compileExpr(e.Rhs)
		return ctx.NewAdd(l, r)
	case *ELessThan:
		l, r := ctx.compileExpr(e.Lhs), ctx.compileExpr(e.Rhs)
		return ctx.NewICmp(enum.IPredSLT, l, r)
	case EConstant:
		return compileConstant(e)
	}
	panic("unimplemented expression")
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
		EConstant
		Stmt
	}
	DefaultCase Stmt
}
type SDoWhile struct {
	Stmt
	Cond  Expr
	Block Stmt
}
type SForLoop struct {
	Stmt
	InitName string
	InitExpr Expr
	Step     Expr
	Cond     Expr
	Block    Stmt
}
type SWhile struct {
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

type Context struct {
	*extend.ExtBlock
	parent     *Context
	vars       map[string]value.Value
	leaveBlock *ir.Block
}

func NewContext(b *ir.Block) *Context {
	return &Context{
		ExtBlock:   extend.Block(b),
		parent:     nil,
		vars:       make(map[string]value.Value),
		leaveBlock: nil,
	}
}

func (c *Context) NewContext(b *ir.Block) *Context {
	ctx := NewContext(b)
	ctx.parent = c
	return ctx
}

func (c Context) lookupVariable(name string) value.Value {
	if v, ok := c.vars[name]; ok {
		return v
	} else if c.parent != nil {
		return c.parent.lookupVariable(name)
	} else {
		fmt.Printf("variable: `%s`\n", name)
		panic("no such variable")
	}
}

func (ctx *Context) compileStmt(stmt Stmt) {
	if !ctx.BelongsToFunc() {
		return
	}
	f := ctx.Parent
	switch s := stmt.(type) {
	case *SIf:
		thenCtx := ctx.NewContext(f.NewBlock("if.then"))
		thenCtx.compileStmt(s.Then)
		elseCtx := ctx.NewContext(f.NewBlock("if.else"))
		elseCtx.compileStmt(s.Else)
		ctx.NewCondBr(ctx.compileExpr(s.Cond), thenCtx.Block, elseCtx.Block)
		if !thenCtx.HasTerminator() {
			leaveB := f.NewBlock("leave.if")
			thenCtx.NewBr(leaveB)
		}
	case *SSwitch:
		cases := []*ir.Case{}
		for _, ca := range s.CaseList {
			caseB := f.NewBlock("switch.case")
			ctx.NewContext(caseB).compileStmt(ca.Stmt)
			cases = append(cases, ir.NewCase(compileConstant(ca.EConstant), caseB))
		}
		defaultB := f.NewBlock("switch.default")
		ctx.NewContext(defaultB).compileStmt(s.DefaultCase)
		ctx.NewSwitch(ctx.compileExpr(s.Target), defaultB, cases...)
	case *SDoWhile:
		doCtx := ctx.NewContext(f.NewBlock("do.while.body"))
		ctx.NewBr(doCtx.Block)
		leaveB := f.NewBlock("leave.do.while")
		doCtx.leaveBlock = leaveB
		doCtx.compileStmt(s.Block)
		doCtx.NewCondBr(doCtx.compileExpr(s.Cond), doCtx.Block, leaveB)
	case *SForLoop:
		loopCtx := ctx.NewContext(f.NewBlock("for.loop.body"))
		ctx.NewBr(loopCtx.Block)
		firstAppear := loopCtx.NewPhi(ir.NewIncoming(loopCtx.compileExpr(s.InitExpr), ctx.Block))
		loopCtx.vars[s.InitName] = firstAppear
		step := loopCtx.compileExpr(s.Step)
		firstAppear.Incs = append(firstAppear.Incs, ir.NewIncoming(step, loopCtx.Block))
		loopCtx.vars[s.InitName] = step
		leaveB := f.NewBlock("leave.for.loop")
		loopCtx.leaveBlock = leaveB
		loopCtx.compileStmt(s.Block)
		loopCtx.NewCondBr(loopCtx.compileExpr(s.Cond), loopCtx.Block, leaveB)
	case *SWhile:
		condCtx := ctx.NewContext(f.NewBlock("while.loop.cond"))
		ctx.NewBr(condCtx.Block)
		loopCtx := ctx.NewContext(f.NewBlock("while.loop.body"))
		leaveB := f.NewBlock("leave.do.while")
		condCtx.NewCondBr(condCtx.compileExpr(s.Cond), loopCtx.Block, leaveB)
		condCtx.leaveBlock = leaveB
		loopCtx.leaveBlock = leaveB
		loopCtx.compileStmt(s.Block)
		loopCtx.NewBr(condCtx.Block)
	case *SDefine:
		v := ctx.NewAlloca(s.Typ)
		ctx.NewStore(ctx.compileExpr(s.Expr), v)
		ctx.vars[s.Name] = v
	case *SRet:
		ctx.NewRet(ctx.compileExpr(s.Val))
	case *SBreak:
		ctx.NewBr(ctx.leaveBlock)
	}
}
