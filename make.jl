using Documenter

makedocs(
    sitename = "llir/llvm",
    format = Documenter.HTML(),
    pages = [
        "index.md",
        "User Guide" => [
            "user-guide/support.md",
            "user-guide/basic.md",
            "user-guide/control.md",
            "user-guide/funcs.md",
            "user-guide/types.md"
        ]
    ]
)

# Documenter can also automatically deploy documentation to gh-pages.
# See "Hosting Documentation" and deploydocs() in the Documenter manual
# for more information.
#=deploydocs(
    repo = "<repository url>"
)=#
