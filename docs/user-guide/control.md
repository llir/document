# Control Flow

Before we start, we need to prepare compile function for something like **expression** and **statement** that not our target.

```go
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
```

And compile functions:

```go
func compileConstant(e EConstant) constant.Constant {
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
	case EConstant:
		return compileConstant(e)
	}
	panic("unimplemented expression")
}
```

`EVariable` would need context to remember variable's value. Here is the related definition of `Context`:

```go
type Context struct {
	*ir.Block
	parent *Context
	vars   map[string]value.Value
}

func NewContext(b *ir.Block) *Context {
	return &Context{
		Block: b,
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
```

Finally, we would have some simple statement as placeholder:

```go
type Stmt interface{ isStmt() Stmt }
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
```

Then compile:

```go
func (ctx *Context) compileStmt(stmt Stmt) {
	if ctx.Parent != nil {
		return
	}
	f := ctx.Parent
	switch s := stmt.(type) {
	case *SDefine:
		v := ctx.NewAlloca(s.Typ)
		ctx.NewStore(ctx.compileExpr(s.Expr), v)
		ctx.vars[s.Name] = v
	case *SRet:
		ctx.NewRet(ctx.compileExpr(s.Val))
	}
}
```

### If

Since we can let:

```go
if condition {
    // A
} else if condition {
    // B
} else {
    // C
}
```

became:

```go
if condition {
    // A
} else {
    if condition {
        // B
    } else {
        // C
    }
}
```

We don't have to convert any **else-if** pattern. Therefore, our source AST looks like this:

```go
type Expr interface{ isExpr() Expr }
type EVoid struct{ Expr }
type EBool struct {
	Expr
	V bool
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
```

First, we limit expression to `EBool` and `EVoid`, and statement to `SIf` and `SRet`, to get a simple subset to focus on our purpose. Then we can get transformers to generate control flow **if**.

1. generate value for constant, `0` for `false`, non `0` for `true`, `void` is `nil` in llir/llvm.

```go
func compileConstant(b *ir.Block, e Expr) constant.Constant {
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
```

2. use **conditional jump** to generate **if** statement

```go
func compileStmt(bb *ir.Block, stmt Stmt) {
    f := bb.Parent
    switch s := stmt.(type) {
    case *SIf:
    	thenB := f.NewBlock("")
    	compileStmt(f, thenB, s.Then)
    	elseB := f.NewBlock("")
    	compileStmt(f, elseB, s.Else)
    	bb.NewCondBr(compileConstant(bb, s.Cond), thenB, elseB)
    	if thenB.Term == nil {
    		leaveB := f.NewBlock("")
    		thenB.NewBr(leaveB)
    	}
    case *SRet:
    	bb.NewRet(compileConstant(bb, s.Val))
    }
}
```

When generating **if**, the most important thing is **leave block**, when if-then block complete, a jump to skip else block required since there has no **block** in high-level language liked concept in LLVM IR. At the end of a basic-block can be a return and since return would terminate a function, jump after return is a dead code, so we have to check we have to generate **leave block** or not. Here is a small example as usage:

```go
f := ir.NewFunc("foo", types.Void)
bb := f.NewBlock("")

compileStmt(f, bb, &SIf{
    Cond: &EBool{V: true},
    Then: nil,
    Else: &SRet{Val: &EVoid{}},
})

// whatever what we did in compileStmt, we use convention that a block leave in the end is empty.
f.Blocks[len(f.Blocks)-1].NewRet(nil)
```

We didn't support else-if directly at here, then we need to know how to handle this via parsing. First, we handle a sequence of `if` `(` `<expr>` `)` `<block>`. Ok, we can fill AST with `Cond` and `Then`, now we should get a token `else`, then we expect a `<block>` or `if`. When we get a `<block>` this is a obviously can be use as `Else`, else a `if` we keep parsing and use it as `Else` statement since `if` for sure is a statement. Of course, with this method, generated IR would have some useless label and jump, but flow analyzing should optimize them later, so it's fine.

### Switch

LLVM has [switch instruction](https://llvm.org/docs/LangRef.html#switch-instruction), hence, we can use it directly.

```go
type SSwitch struct {
    Stmt
    Target   Expr
    CaseList []struct {
        Expr
        Stmt
    }
    DefaultCase Stmt
}

func compileStmt(bb *ir.Block, stmt Stmt) {
    switch s := stmt.(type) {
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
    }
}
```

For every case, we generate a block, then we can jump to target. Then we put statements into case blocks. Finally, we generate switch for the input block. Notice that, switch instruction of LLVM won't generate `break` automatically, you can use the same trick in the previous section **If** to generate auto leave block for each case(Go semantic), or record leave block and introduces break statement(C semantic). Now let's test it:

```go
f := ir.NewFunc("foo", types.Void)
b := f.NewBlock("")

compileStmt(b, &SSwitch{
    Target: &EBool{V: true},
    CaseList: []struct {
        Expr
        Stmt
    }{
        {Expr: &EBool{V: true}, Stmt: &SRet{Val: &EVoid{}}},
    },
    DefaultCase: &SRet{Val: &EVoid{}},
})
```

The switch statement in this section is quite naive, for advanced semantic like pattern matching with extraction or where clause, you would need to do more.

### Loop

#### Do While

Do while is the simplest loop structure since it's code structure almost same to the IR structure. Here we go:

```go
type SDoWhile struct {
	Stmt
	Cond  Expr
	Block Stmt
}

func compileStmt(block *ir.Block, stmt Stmt) {
    switch s := stmt.(type) {
    case *SDoWhile:
        doB := b.Block
        // if previous block is not empty, then we need to create new block for do-while loop
        if b.Insts != nil {
            doB = f.NewBlock("")
        }
        compileStmt(doB, s.Block)
        leaveB := f.NewBlock("")
        doB.NewCondBr(compileConstant(s.Cond), doB, leaveB)
    }
}
```

Can see that, first we check last block is empty or not, if it's empty we keep using it as do-while body. Then we have a leave block, in the end of the do-while body we jump out to leave block or body again depends on condition. Let's test it:

```go
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
```

`SDefine` is define for helping since terminator like `ret` would be remove when we put another terminator `condbr`, and return in a loop as an example is weird. Here is code generation for `SDefine`:

```go
type SDefine struct {
	Stmt
	Name string
	Typ  types.Type
	Expr Expr
}

func compileStmt(block *ir.Block, stmt Stmt) {
    switch s := stmt.(type) {
    case *SDefine:
        v := b.NewAlloca(s.Typ)
        v.SetName(s.Name)
        b.NewStore(compileConstant(s.Expr), v)
    }
}
```
