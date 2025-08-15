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