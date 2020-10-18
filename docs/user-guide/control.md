# Control Flow

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
func compileStmt(bb *ir.Block, stmt Stmt) {
    switch s := stmt.(type) {
    /// ignore
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

For every case, we generate a block, then we can jump to target. Then we put statements into case blocks. Finally, we generate switch for the input block. Notice that, switch instruction of LLVM won't generate `break` automatically, you can use the same trick in previous section **If** to generate auto leave block for each case(Go semantic), or record leave block and introduces break statement(C semantic). Now let's test it:

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
