# Welcome to llir/llvm

## Overview

#### Why LLVM?

When creating a compiler, a classical design looks like this:

{% dot attack_plan.svg
digraph hierarchy {
  node [color=Black,fontname=Courier,shape=box] //All nodes will this shape and colour

 "Source Code"->Frontend->Optimizer->Backend->"Machine Code"
}
%}

This is quite good in old days. There has only one input language, and one target machine.

But there has more and more target machines have to support! Therefore, we need LLVM. Here is the new design:

{% dot attack_plan.svg
digraph hierarchy {
  nodesep=1.0 // increases the separation between nodes

  node [color=Black,fontname=Courier,shape=box] //All nodes will this shape and colour

 {"C Frontend" "Fortran Frontend" "Ada Frontend"}->Optimizer->{"X86 Backend" "PowerPC Backend" "ARM Backend"}
}
%}

Now we only have to focus on our frontend and optimizer! Thanks you, Chris Lattner and who had work for LLVM.

#### Why llir/llvm?

The target of [llir/llvm](https://github.com/llir/llvm) is: interact in Go with LLVM IR without binding with LLVM.
Therefore, you don't have to compile LLVM(could take few hours), no fighting with cgo.
Working under pure Go environment and start your journey.

## Installation

`go get github.com/llir/llvm`
