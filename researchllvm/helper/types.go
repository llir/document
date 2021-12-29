package helper

import "github.com/llir/llvm/ir/types"

var (
	TVoid = types.Void
	TI8   = types.I8
	TI32  = types.I32
	TI64  = types.I64
)

func TPtr(t types.Type) *types.PointerType {
	return types.NewPointer(t)
}
