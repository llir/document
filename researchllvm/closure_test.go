package researchllvm

import (
	"testing"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

func TestClosure(t *testing.T) {
	m := ir.NewModule()

	zero := constant.NewInt(types.I32, 0)
	one := constant.NewInt(types.I32, 1)
	printf := PrintfPlugin(m)

	captureStruct := m.NewTypeDef("id_capture", types.NewStruct(
		types.I32,
	))
	captureTyp := types.NewPointer(captureStruct)
	idFn := m.NewFunc("id", types.I32, ir.NewParam("capture", captureTyp))
	idB := idFn.NewBlock("")
	v := idB.NewGetElementPtr(captureStruct, idFn.Params[0], zero, zero)
	idB.NewRet(idB.NewLoad(types.I32, v))
	idClosureTyp := m.NewTypeDef("id_closure", types.NewStruct(
		captureTyp,
		idFn.Type(),
	))

	mainFn := m.NewFunc("main", types.I32)
	b := mainFn.NewBlock("")
	// define a local variable `i`
	i := b.NewAlloca(types.I32)
	b.NewStore(constant.NewInt(types.I32, 10), i)
	// use alloca at here to simplify code, in real case should be `malloc` or `gc_malloc`
	captureInstance := b.NewAlloca(captureStruct)
	ptrToCapture := b.NewGetElementPtr(captureStruct, captureInstance, zero, zero)
	// capture variable
	b.NewStore(b.NewLoad(types.I32, i), ptrToCapture)
	// prepare closure
	idClosure := b.NewAlloca(idClosureTyp)
	ptrToCapturePtr := b.NewGetElementPtr(idClosureTyp, idClosure, zero, zero)
	b.NewStore(captureInstance, ptrToCapturePtr)
	ptrToFuncPtr := b.NewGetElementPtr(idClosureTyp, idClosure, zero, one)
	b.NewStore(idFn, ptrToFuncPtr)
	// assuming we transfer closure into another context
	accessCapture := b.NewGetElementPtr(idClosureTyp, idClosure, zero, zero)
	accessFunc := b.NewGetElementPtr(idClosureTyp, idClosure, zero, one)
	result := b.NewCall(b.NewLoad(idFn.Type(), accessFunc), b.NewLoad(captureTyp, accessCapture))

	printIntegerFormat := m.NewGlobalDef("tmp", constant.NewCharArrayFromString("%d\n"))
	pointerToString := b.NewGetElementPtr(types.NewArray(3, types.I8), printIntegerFormat, zero, zero)
	b.NewCall(printf, pointerToString, result)

	b.NewRet(constant.NewInt(types.I32, 0))

	PrettyPrint(m)

	executeIR(m)
}

// generated LLVM IR:
//
// ```
// %id_capture = type { i32 }
// %id_closure = type { %id_capture*, i32 (%id_capture*)* }
//
// @tmp = global [3 x i8] c"%d\0A"
//
// declare i32 @printf(i8* %format, ...)
//
// define i32 @id(%id_capture* %capture) {
// ; <label>:0
// 	%1 = getelementptr %id_capture, %id_capture* %capture, i32 0, i32 0
// 	%2 = load i32, i32* %1
// 	ret i32 %2
// }
//
// define i32 @main() {
// ; <label>:0
// 	 %1 = alloca i32
// 	 store i32 10, i32* %1
// 	 %2 = alloca %id_capture
// 	 %3 = getelementptr %id_capture, %id_capture* %2, i32 0, i32 0
// 	 %4 = load i32, i32* %1
// 	 store i32 %4, i32* %3
// 	 %5 = alloca %id_closure
// 	 %6 = getelementptr %id_closure, %id_closure* %5, i32 0, i32 0
// 	 store %id_capture* %2, %id_capture** %6
// 	 %7 = getelementptr %id_closure, %id_closure* %5, i32 0, i32 1
// 	 store i32 (%id_capture*)* @id, i32 (%id_capture*)** %7
// 	 %8 = getelementptr %id_closure, %id_closure* %5, i32 0, i32 0
// 	 %9 = getelementptr %id_closure, %id_closure* %5, i32 0, i32 1
// 	 %10 = load i32 (%id_capture*)*, i32 (%id_capture*)** %9
// 	 %11 = load %id_capture*, %id_capture** %8
// 	 %12 = call i32 %10(%id_capture* %11)
// 	 %13 = getelementptr [3 x i8], [3 x i8]* @tmp, i32 0, i32 0
// 	 %14 = call i32 (i8*, ...) @printf(i8* %13, i32 %12)
// 	 ret i32 0
// }
// ```
// Output:
//
// 10
