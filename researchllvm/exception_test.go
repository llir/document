package researchllvm

import (
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	. "github.com/llir/researchllvm/helper"
)

func TestException(t *testing.T) {
	m := ir.NewModule()

	_ZTIi := m.NewGlobal("_ZTIi", TPtr(TPtr(TI8)))
	_ZTIi.Linkage = enum.LinkageExternal
	__cxa_allocate_exception := m.NewFunc("__cxa_allocate_exception", TPtr(TI8),
		ir.NewParam("", TI64),
	)
	__cxa_throw := m.NewFunc("__cxa_throw", TVoid,
		ir.NewParam("exception_header", TPtr(TI8)),
		ir.NewParam("", TPtr(TI8)),
		ir.NewParam("", TPtr(TI8)),
	)

	exceptionThrower := m.NewFunc("I throw exception!", TVoid)
	bb := exceptionThrower.NewBlock("")
	exception_header := bb.NewCall(__cxa_allocate_exception, CI64(4))
	payload := bb.NewBitCast(exception_header, TPtr(TI32))
	bb.NewStore(CI32(1), payload)
	bb.NewCall(__cxa_throw,
		exception_header,
		bb.NewBitCast(_ZTIi, TPtr(TI8)),
		constant.NewNull(TPtr(TI8)),
	)
	bb.NewRet(nil)

	main := m.NewFunc(
		"main",
		TI32,
	)
	mainB := main.NewBlock("")
	mainB.NewCall(exceptionThrower)
	mainB.NewRet(CI32(0))

	PrettyPrint(m)

	ExecuteIR(m)
}
