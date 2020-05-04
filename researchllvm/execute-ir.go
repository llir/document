package researchllvm

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/llir/llvm/ir"
)

func executeIR(mod *ir.Module) {
	tmpIRName := "tmp.ll"
	tmpIR, err := os.Create(tmpIRName)
	if err != nil {
		panic(err)
	}
	_, err = mod.WriteTo(tmpIR)
	if err != nil {
		panic(err)
	}
	cmd := exec.Command("lli", "tmp.ll")
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Output:\n\n%s\n", stdoutStderr)
	err = os.Remove(tmpIRName)
	if err != nil {
		panic(err)
	}
}
