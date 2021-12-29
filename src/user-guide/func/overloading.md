# Function Overloading

There is no overloading in LLVM IR. One solution is to create one function per function signature, where each LLVM IR function would have a unique name (this is why C++ compilers do name mangling).
