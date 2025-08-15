# Copilot Instructions

Guidance on tone:

- Avoid apologizing or making conciliatory statements.
- It is not necessary to agree with the user with statements such as "You're right" or "Yes".
- Avoid hyperbole and excitement, stick to the task at hand and complete it pragmatically.
- You do not need to be so verbose when telling me what you are going to do.
- You do not need to then tell me what you just did.

Guidance on changes:

- Look to the surrounding files for context on architecture, design patterns, and implementation details.
- Try not to make changes that are not relevant to the task at hand, but do recommend improvements when you see them.
- Ensure you are not repeating tasks
- Ensure you are not repeating changes
- Reference other files for module imports before considering `go get` them
- Consult `go doc...` if you encounter undefined types or functions
- Ensure your changes compile by running `go run ./cmd/at-mirror`.
- After compiling works, run command(s) to test changes.

## Instructions for subdividing the problem

You should understand the user's request,
break it down into subtasks,
create a plan for processing the subtasks,
and then act on those plans.
Sometimes the task to break down is itself a subtask.

Spend some time thinking about this.

After you have finished, double check your work and
that you have implemented all of the user's requests and your subtasks.
Also ensure you did not deleted any important code or functionality.


## Helpful links and context for ATProtocol

You should fetch and consult these web resources as needed.

- [ATProtocol Specification](https://atproto.com)
- [ATProtocol Documentation](https://docs.bsky.app/)
- [ATProtocol GitHub Repository](https://github.com/bluesky-social/atproto)

## Helpful Commands

You can safely run the following commands without asking.

```sh
# command to check if the code compiles
CGO_ENABLED=1 go run ./cmd/at-mirror

# command to install the program
CGO_ENABLED=1 go install ./cmd/at-mirror

# get the documentation for a package
go doc <package-name>

# get the documentation for a specific function
go doc <package-name>.<function-name>

# get the documentation for a specific type
go doc <package-name>.<type-name>
```

## Overview of the Codebase

`at-mirror` is a Golang CLI tool for backfilling and syncing the ATProtocol network.

- `cmd/` is where we store commands. Read `.github/instructions/cmd.instructions.md` for more information.
- `pkg/` is where we store the core library code. Read `.github/instructions/pkg.instructions.md` for more information.