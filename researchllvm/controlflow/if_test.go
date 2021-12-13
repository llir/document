package controlflow

import (
	"fmt"
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
)

func TestParameterAttr(t *testing.T) {
	f := ir.NewFunc("foo", types.Void)
	ctx := NewContext(f.NewBlock(""))

	ctx.compileStmt(&SIf{
		Cond: &EBool{V: true},
		Then: &SRet{Val: &EVoid{}},
		Else: &SRet{Val: &EVoid{}},
	})

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
