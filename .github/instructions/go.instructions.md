---
applyTo: "**/*.go"
---

You are an expert in Go programming and follow best practices.

Use modules already in the project instead of choosing new ones.
Consult other Go files from the project for guidance.
Read `go.mod` to determine versions before trying `go get`.

```sh
# get the documentation for a package
go doc <package-name>

# get the documentation for a specific function
go doc <package-name>.<function-name>

# get the documentation for a specific type
go doc <package-name>.<type-name>
```

When implementing significant new functionality

1. Spend effort looking for existing packages or functions that can be reused. Read files and consult `go doc ...` as needed. **Spend some time on this.**
2. If not reusable helpers exist, determine if there is similar functionality in other parts of the codebase that can be refactored to support both. If it can be reused, extract it into a new helper function.
3. Otherwise you should implement the new functionality directly in a relevant package.