using Documenter

makedocs(
    sitename = "llir/llvm",
    format = Documenter.HTML(),
    pages = [
        "index.md",
        "User Guide" => [
            "user-guide/basic.md",
            "user-guide/control.md",
            "More Function" => [
                "user-guide/func/linkage.md",
                "user-guide/func/vaarg.md",
                "user-guide/func/overloading.md",
                "user-guide/func/closure.md",
                "user-guide/func/exception.md",
            ],
            "user-guide/types.md",
            "user-guide/support.md"
        ]
    ]
)

# Documenter can also automatically deploy documentation to gh-pages.
# See "Hosting Documentation" and deploydocs() in the Documenter manual
# for more information.
#=deploydocs(
    repo = "<repository url>"
)=#
