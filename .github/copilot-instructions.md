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

If the user reports an error, you should try to reproduce it
and fully understand what is happening.

Spend time thinking.

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

# command to get help for a subcommand
CGO_ENABLED=1 go run ./cmd/at-mirror <cmd> [subcmd] --help
```

`verdverm.com` is a good account to test commands with

Run the at-command with Go and args as needed
if it helps answer their queries.


## Overview of the Codebase

`at-mirror` is a Golang CLI tool and set of packages for

1. backfilling and syncing the ATProtocol network.
2. extracting info from repos and expanding the dataset.
3. running ai models on repo and record data.
4. composing feeds from the extracted and enriched data.


- `cmd/` is where we store CLI command code. Read `.github/instructions/cmd.instructions.md` for more context.
- `pkg/` is where we store the core library code. Read `.github/instructions/pkg.instructions.md` for more context.
- If the user talks about SQL related topics, read `.github/instructions/sql.instructions.md` for more context.
- If the user talks about Go or you read other Go files, read `.github/instructions/go.instructions.md` for more context.