---
mode: 'agent'
tools: ['codebase', 'usages', 'vscodeAPI', 'problems', 'changes', 'testFailure', 'terminalSelection', 'terminalLastCommand', 'fetch', 'findTestFiles', 'searchResults', 'githubRepo', 'extensions', 'runTests', 'editFiles', 'search', 'new', 'runCommands', 'runTasks']
---

# Create a Go command '${input:usage}'

[go-command-structure](../snippets/go-cmd-structure.md)
[go-common][../instructions/go.instructions.md]

## Guidance on implementing a new command

Follow the file layout and naming conventions
which follow the command hierarchy.
Before deciding how to implement the new command,
look up several other commands.
Select examples both near the new command file
and those that seem related.

When implementing the new command

- follow the patterns and top-level logic of existing commands
- reusable logic is implemented in `${workspaceFolder}/pkg/...`

${input}