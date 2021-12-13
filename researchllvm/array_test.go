package researchllvm

import (
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	. "github.com/llir/researchllvm/helper"
)

var formatString = "array_def[%d]: %d\n"

func TestArray(t *testing.T) {
	mod := ir.NewModule()

	arrTy := types.NewArray(5, types.I8)
	arrayDef := mod.NewGlobalDef("array_def", constant.NewArray(arrTy, CI8(1), CI8(2), CI8(3), CI8(4), CI8(5)))

	printf := PrintfPlugin(mod)

	fmtStr := mod.NewGlobalDef("x", constant.NewCharArrayFromString(formatString))
	main := mod.NewFunc("main", types.I32)
	mainB := main.NewBlock("")
	ptrToStr := mainB.NewGetElementPtr(
		types.NewArray(uint64(len(formatString)), types.I8), fmtStr,
		CI32(0), CI32(0),
	)
	arr := mainB.NewLoad(arrTy, arrayDef)
	for i := 0; i < 5; i++ {
		mainB.NewCall(printf, ptrToStr, CI32(int64(i)), mainB.NewExtractValue(arr, uint64(i)))
		mainB.NewInsertValue(arr, CI8(0), uint64(i))
		mainB.NewCall(printf, ptrToStr, CI32(int64(i)), mainB.NewExtractValue(arr, uint64(i)))
	}
	for i := 0; i < 5; i++ {
		pToElem := mainB.NewGetElementPtr(arrTy, arrayDef, CI32(0), CI32(int64(i)))
		mainB.NewCall(printf, ptrToStr, CI32(int64(i)),
			mainB.NewLoad(types.I8, pToElem))
		mainB.NewStore(CI8(0), pToElem)
		mainB.NewCall(printf, ptrToStr, CI32(int64(i)),
			mainB.NewLoad(types.I8, pToElem))
	}
	for i := 0; i < 5; i++ {
		mainB.NewCall(printf, ptrToStr, CI32(int64(i)), mainB.NewExtractValue(arr, uint64(i)))
		newArr := mainB.NewInsertValue(arr, CI8(0), uint64(i))
		mainB.NewCall(printf, ptrToStr, CI32(int64(i)), mainB.NewExtractValue(newArr, uint64(i)))
	}
	mainB.NewRet(CI32(0))

	PrettyPrint(mod)

	ExecuteIR(mod)
}
