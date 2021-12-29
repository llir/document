# Exception

In this section, you will see how to reuse c++ exception in LLVM.

First, we need to setup a set of function from c++ ABI:

```go
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
```

And a helper for throw exception from a block:

```go
func throwException(m *ModuleWithException, bb *ir.Block) {
    // C++ requires one allocate an exception first
	payload := bb.NewCall(m.__cxa_allocate_exception, CI64(4))
    // now we stores I32 `1` into payload
	bb.NewStore(CI32(1), bb.NewBitCast(payload, TPtr(TI32)))
    // finally, we call `__cxa_throw` to throw exception
	bb.NewCall(m.__cxa_throw,
		payload,
		constant.NewBitCast(m._ZTIi, TPtr(TI8)),
		constant.NewNull(TPtr(TI8)),
	)
}
```

Finally, is our full example

```go
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
// we must use invoke when a function might throw exception, we need to give
// 1. normal return block for function returns normally
// 2. exception return block for function throws an exception
mainB.NewInvoke(exceptionThrower, []value.Value{}, normalRetB, exceptionRetB)
normalRetB.NewRet(CI32(0))
// landingpad stands for catch and cleanup
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
catchintB := main.NewBlock("catchint")
cleanupB := main.NewBlock("cleanup")
cleanupB.NewResume(exc)
// we check typeinfo is as expected
// 1. if type info is same as our expection, we going to catchint block to handle threw I32
// 2. if not, we cleanup exception
tst_int := exceptionRetB.NewICmp(enum.IPredEQ, exc_sel, tid_int)
exceptionRetB.NewCondBr(tst_int, catchintB, cleanupB)
// in catchint block, we
// 1. call `__cxa_begin_catch` to begin catching
// 2. load payload to get threw I32
// 3. call `__cxa_end_catch` to end catching
payload := catchintB.NewCall(m.__cxa_begin_catch, exc_ptr)
payload.Tail = enum.TailTail
payload_int := catchintB.NewBitCast(payload, TPtr(TI32))
retval := catchintB.NewLoad(TI32, payload_int)
end_catch := catchintB.NewCall(m.__cxa_end_catch)
end_catch.Tail = enum.TailTail

// Finally, we return threw value from main
returnB := main.NewBlock("return")
catchintB.NewBr(returnB)
returnB.NewRet(retval)
```
