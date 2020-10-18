package controlflow

import (
	"fmt"
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/types"
)

func TestParameterAttr(t *testing.T) {
	f := ir.NewFunc("foo", types.Void)
	bb := f.NewBlock("")

	compileStmt(bb, &SIf{
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
