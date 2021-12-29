package researchllvm

import (
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
	. "github.com/llir/researchllvm/helper"
)

type ModuleWithException struct {
	*ir.Module
	_ZTIi                    *ir.Global
	__cxa_allocate_exception *ir.Func
	__cxa_throw              *ir.Func
	__cxa_begin_catch        *ir.Func
	__cxa_end_catch          *ir.Func
	__cxa_call_unexpected    *ir.Func
	llvm_eh_typeid_for       *ir.Func
}

func NewModuleWithException() *ModuleWithException {
	m := ir.NewModule()
	mWithE := &ModuleWithException{
		Module: m,
		_ZTIi:  m.NewGlobal("_ZTIi", TPtr(TI8)),
		__cxa_allocate_exception: m.NewFunc("__cxa_allocate_exception", TPtr(TI8),
			ir.NewParam("", TI64),
		),
		__cxa_throw: m.NewFunc("__cxa_throw", TVoid,
			ir.NewParam("exception_header", TPtr(TI8)),
			ir.NewParam("", TPtr(TI8)),
			ir.NewParam("", TPtr(TI8)),
		),
		__cxa_begin_catch:     m.NewFunc("__cxa_begin_catch", TPtr(TI8), ir.NewParam("", TPtr(TI8))),
		__cxa_end_catch:       m.NewFunc("__cxa_end_catch", TVoid),
		__cxa_call_unexpected: m.NewFunc("__cxa_call_unexpected", TVoid, ir.NewParam("", TPtr(TI8))),
		llvm_eh_typeid_for:    m.NewFunc("llvm.eh.typeid.for", TI32, ir.NewParam("", TPtr(TI8))),
	}
	mWithE._ZTIi.Linkage = enum.LinkageExternal
	return mWithE
}

func throwException(m *ModuleWithException, bb *ir.Block) {
	exception_header := bb.NewCall(m.__cxa_allocate_exception, CI64(4))
	bb.NewStore(CI32(1), bb.NewBitCast(exception_header, TPtr(TI32)))
	bb.NewCall(m.__cxa_throw,
		exception_header,
		constant.NewBitCast(m._ZTIi, TPtr(TI8)),
		constant.NewNull(TPtr(TI8)),
	)
}

func TestException(t *testing.T) {
	m := NewModuleWithException()

	exceptionThrower := m.NewFunc("I throw exception!", TI32)
	bb := exceptionThrower.NewBlock("")
	throwException(m, bb)
	bb.NewRet(CI32(1))

	main := m.NewFunc("main", TI32)
	main.Personality = constant.True
	mainB := main.NewBlock("")
	normalRetB := main.NewBlock("normalRet")
	exceptionRetB := main.NewBlock("exceptionRet")
	mainB.NewInvoke(exceptionThrower, []value.Value{}, normalRetB, exceptionRetB)
	normalRetB.NewRet(CI32(0))
	exc := exceptionRetB.NewLandingPad(types.NewStruct(TPtr(TI8), TI32),
		ir.NewClause(enum.ClauseTypeCatch, constant.NewBitCast(m._ZTIi, TPtr(TI8))),
		ir.NewClause(enum.ClauseTypeFilter,
			constant.NewArray(types.NewArray(1, TPtr(TI8)), constant.NewBitCast(m._ZTIi, TPtr(TI8))),
		),
	)
	exc.Cleanup = true
	exc_ptr := exceptionRetB.NewExtractValue(exc, 0)
	exc_sel := exceptionRetB.NewExtractValue(exc, 1)
	tid_int := exceptionRetB.NewCall(m.llvm_eh_typeid_for, constant.NewBitCast(m._ZTIi, TPtr(TI8)))
	tid_int.Tail = enum.TailTail
	tst_int := exceptionRetB.NewICmp(enum.IPredEQ, exc_sel, tid_int)
	catchintB := main.NewBlock("catchint")
	cleanupB := main.NewBlock("cleanup")
	cleanupB.NewResume(exc)
	exceptionRetB.NewCondBr(tst_int, catchintB, cleanupB)
	payload := catchintB.NewCall(m.__cxa_begin_catch, exc_ptr)
	payload.Tail = enum.TailTail
	payload_int := catchintB.NewBitCast(payload, TPtr(TI32))
	retval := catchintB.NewLoad(TI32, payload_int)
	end_catch := catchintB.NewCall(m.__cxa_end_catch)
	end_catch.Tail = enum.TailTail

	returnB := main.NewBlock("return")
	catchintB.NewBr(returnB)
	returnB.NewRet(retval)

	PrettyPrint(m.Module)

	ExecuteIR(m.Module)
}
