package researchllvm

import (
	"fmt"

	"github.com/llir/llvm/ir"
)

func PrettyPrint(mod *ir.Module) {
	fmt.Printf("generated LLVM IR:\n\n```\n%s```\n", mod)
}
