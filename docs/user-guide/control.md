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
		// we have no boolean in LLVM IR
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

We don't have to convert any **else-if** pattern. Therefore, our `If` looks like this:

```go
type SIf struct {
	Stmt
	Cond Expr
	Then Stmt
	Else Stmt
}
```

Then we can get transformers to generate control flow **if**. Using **conditional jump** to generate **if** statement:

```go
func (ctx *Context) compileStmt(stmt Stmt) {
	switch s := stmt.(type) {
	case *SIf:
		thenCtx := ctx.NewContext(f.NewBlock("if.then"))
		thenCtx.compileStmt(s.Then)
		elseB := f.NewBlock("if.else")
		ctx.NewContext(elseB).compileStmt(s.Else)
		ctx.NewCondBr(ctx.compileExpr(s.Cond), thenCtx.Block, elseB)
		if !thenCtx.HasTerminator() {
			leaveB := f.NewBlock("leave.if")
			thenCtx.NewBr(leaveB)
		}
	}
}
```

When generating **if**, the most important thing is **leave block**, when if-then block complete, a jump to skip else block required since there has no **block** in high-level language liked concept in LLVM IR. At the end of a basic-block can be a return and since return would terminate a function, jump after return is a dead code, so we have to check we have to generate **leave block** or not. Here is a small example as usage:

```go
f := ir.NewFunc("foo", types.Void)
bb := f.NewBlock("")

ctx.compileStmt(&SIf{
    Cond: &EBool{V: true},
    Then: &SRet{Val: &EVoid{}},
    Else: &SRet{Val: &EVoid{}},
})
```

Finally, we get:

```llvm
define void @foo() {
0:
	br i1 true, label %if.then, label %if.else

if.then:
	ret void

if.else:
	ret void
}
```

We didn't support else-if directly at here, then we need to know how to handle this via parsing. First, we handle a sequence of `if` `(` `<expr>` `)` `<block>`. Ok, we can fill AST with `Cond` and `Then`, now we should get a token `else`, then we expect a `<block>` or `if`. When we get a `<block>` this is a obviously can be use as `Else`, else a `if` we keep parsing and use it as `Else` statement since `if` for sure is a statement. Of course, with this method, generated IR would have some useless label and jump, but flow analyzing should optimize them later, so it's fine.

### Switch

LLVM has [switch instruction](https://llvm.org/docs/LangRef.html#switch-instruction), hence, we can use it directly.

```go
type SSwitch struct {
	Stmt
	Target   Expr
	CaseList []struct {
		EConstant // LLVM IR only takes constant, if you want advanced switch semantic, then you can't directly use this approach
		Stmt
	}
	DefaultCase Stmt
}

func (ctx *Context) compileStmt(stmt Stmt) {
	switch s := stmt.(type) {
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
	}
}
```

For every case, we generate a block, then we can jump to target. Then we put statements into case blocks. Finally, we generate switch for the input block. Notice that, switch instruction of LLVM won't generate `break` automatically, you can use the same trick in the previous section **If** to generate auto leave block for each case(Go semantic), or record leave block and introduces break statement(C semantic). Now let's test it:

```go
f := ir.NewFunc("foo", types.Void)
ctx := NewContext(f.NewBlock(""))

ctx.compileStmt(&SSwitch{
	Target: &EBool{V: true},
	CaseList: []struct {
		EConstant
		Stmt
	}{
		{EConstant: &EBool{V: true}, Stmt: &SRet{Val: &EVoid{}}},
	},
	DefaultCase: &SRet{Val: &EVoid{}},
})
```

And output:

```llvm
define void @foo() {
0:
	switch i1 true, label %switch.default [
		i1 true, label %switch.case
	]

switch.case:
	ret void

switch.default:
	ret void
}
```

The switch statement in this section is quite naive, for advanced semantic like pattern matching with extraction or where clause, you would need to do more.

### Loop

#### Break

Break statement needs to extend `Context`, with a new field called `leaveBlock`:

```go
type Context struct {
	// ...
	leaveBlock *ir.Block
}

func NewContext(b *ir.Block) *Context {
	return &Context{
		// ...
		leaveBlock: nil,
	}
}
```

Then it's just a jump:

```go
func (ctx *Context) compileStmt(stmt Stmt) {
	switch s := stmt.(type) {
	case *SBreak:
		ctx.NewBr(ctx.leaveBlock)
	}
}
```

Remember to update leave block information(and remove it when needed), and continue can be done in the same way.

#### Do While

Do while is the simplest loop structure since it's code structure almost same to the IR structure. Here we go:

```go
type SDoWhile struct {
	Stmt
	Cond  Expr
	Block Stmt
}

func (ctx *Context) compileStmt(stmt Stmt) {
	switch s := stmt.(type) {
	case *SDoWhile:
		doCtx := ctx.NewContext(f.NewBlock("do.while.body"))
		ctx.NewBr(doCtx.Block)
		leaveB := f.NewBlock("leave.do.while")
		doCtx.leaveBlock = leaveB
		doCtx.compileStmt(s.Block)
		doCtx.NewCondBr(doCtx.compileExpr(s.Cond), doCtx.Block, leaveB)
	}
}
```

Can see that, we jump to do-while body directly. Then we have a leave block, in the end of the do-while body we jump out to leave block or body again depends on condition. Let's test it:

```go
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
```

And output:

```llvm
define void @foo() {
0:
	br label %do.while.body

do.while.body:
	%1 = alloca i32
	store i32 1, i32* %1
	br i1 true, label %do.while.body, label %leave.do.while

leave.do.while:
	ret void
}
```

#### For Loop

For-loop would be an interesting case, at here, I only present a for-loop that can only have one initialize variable to reduce complexity, therefore, we have a AST like this:

```go
type SForLoop struct {
	Stmt
	InitName string
	InitExpr Expr
	Step     Expr
	Cond     Expr
	Block    Stmt
}
```

For example, `for (x=0; x=x+1; x<10) {}` break down to:

 1.`InitName`: `x`
 2. `InitExpr`: `0`
 3. `Step`: `x+1`
 4. `Cond`: `x<10`
 5. `Block`: `{}`
 
At first view, people might think for-loop is as easy as do-while, but in SSA form, reuse variable in a loop need a new instruction: [phi](https://llvm.org/docs/LangRef.html#i-phi).

```go
func (ctx *Context) compileStmt(stmt Stmt) {
	switch s := stmt.(type) {
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
	}
}
```

1. Create a loop body context
2. jump from the previous block
3. Put phi into loop body
4. Phi would have two incoming, first is `InitExpr`, the second one is `Step` result.
5. compile step
6. compile the conditional branch, jump to loop body or leave block

It generates:

```llvm
define void @foo() {
0:
	br label %for.loop.body

for.loop.body:
	%1 = phi i32 [ 0, %0 ], [ %2, %for.loop.body ]
	%2 = add i32 %1, 1
	%3 = alloca i32
	store i32 2, i32* %3
	%4 = icmp slt i32 %2, 10
	br i1 %4, label %for.loop.body, label %leave.for.loop

leave.for.loop:
	ret void
}
```

In fact, you can also avoid phi, you can make a try as practice.
