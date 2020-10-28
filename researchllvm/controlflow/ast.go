package controlflow

import (
	"fmt"

	"github.com/dannypsnl/extend"
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
type EI32 struct {
	Expr
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
	default:
		return compileConstant(e)
	}
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
type SForLoop struct {
	Stmt
	Init  *SDefine
	Step  Stmt
	Cond  Expr
	Block Stmt
}
type SDefine struct {
	Stmt
	Name string
	Typ  types.Type
	Expr Expr
}
type SAssign struct {
	Stmt
	Name string
	Expr Expr
}
type SRet struct {
	Stmt
	Val Expr
}

type Context struct {
	*extend.ExtBlock
	parent *Context
	vars   map[string]value.Value
}

func NewContext(b *ir.Block) *Context {
	return &Context{
		ExtBlock: extend.Block(b),
		parent:   nil,
		vars:     make(map[string]value.Value),
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
		thenB := extend.Block(f.NewBlock(""))
		ctx.NewContext(thenB.Block).compileStmt(s.Then)
		elseB := f.NewBlock("")
		ctx.NewContext(elseB).compileStmt(s.Else)
		ctx.NewCondBr(ctx.compileExpr(s.Cond), thenB.Block, elseB)
		if thenB.HasTerminator() {
			leaveB := f.NewBlock("")
			thenB.NewBr(leaveB)
		}
	case *SSwitch:
		cases := []*ir.Case{}
		for _, ca := range s.CaseList {
			caseB := f.NewBlock("")
			ctx.NewContext(caseB).compileStmt(ca.Stmt)
			cases = append(cases, ir.NewCase(compileConstant(ca.Expr), caseB))
		}
		defaultB := f.NewBlock("")
		ctx.NewContext(defaultB).compileStmt(s.DefaultCase)
		ctx.NewSwitch(ctx.compileExpr(s.Target), defaultB, cases...)
	case *SDoWhile:
		doB := ctx.Block
		// if previous block is not empty, then we need to create new block for do-while loop
		if ctx.Insts != nil {
			doB = f.NewBlock("")
		}
		doCtx := ctx.NewContext(doB)
		doCtx.compileStmt(s.Block)
		leaveB := f.NewBlock("")
		doB.NewCondBr(doCtx.compileExpr(s.Cond), doB, leaveB)
	case *SForLoop:
		ctx.compileStmt(s.Init)
		loopCtx := ctx.NewContext(f.NewBlock(""))
		leaveB := f.NewBlock("")
		ctx.NewCondBr(ctx.compileExpr(s.Cond), loopCtx.Block, leaveB)
		loopCtx.compileStmt(s.Block)
		loopCtx.compileStmt(s.Step)
		loopCtx.NewCondBr(loopCtx.compileExpr(s.Cond), loopCtx.Block, leaveB)
	case *SDefine:
		v := ctx.NewAlloca(s.Typ)
		v.SetName(s.Name)
		ctx.NewStore(ctx.compileExpr(s.Expr), v)
		ctx.vars[s.Name] = v
	case *SAssign:
		exp := ctx.compileExpr(s.Expr)
		v := ctx.NewAlloca(exp.Type())
		v.SetName(s.Name)
		ctx.NewStore(exp, v)
		ctx.vars[s.Name] = v
	case *SRet:
		ctx.NewRet(ctx.compileExpr(s.Val))
	}
}
