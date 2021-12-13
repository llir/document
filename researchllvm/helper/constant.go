package helper

import (
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
)

func CI8(x int64) constant.Constant  { return constant.NewInt(types.I8, x) }
func CI16(x int64) constant.Constant { return constant.NewInt(types.I16, x) }
func CI32(x int64) constant.Constant { return constant.NewInt(types.I32, x) }
func CI64(x int64) constant.Constant { return constant.NewInt(types.I64, x) }

func CF32(x float64) constant.Constant { return constant.NewFloat(types.Float, x) }
func CF64(x float64) constant.Constant { return constant.NewFloat(types.Double, x) }
