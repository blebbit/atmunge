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
- Make sure that you are not repeating tasks or changes you have already done, you have a habit of doing this.
- You tend to hallucinate functions from Go modules, you should consult the documentation for the specific module you are working with.
- Ensure your changes compile by running `go run ./cmd/at-mirror`.

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

If you add or remove a file, you should update this section accordingly.

```sh
.
├── Dockerfile
├── LICENSE
├── Makefile
├── README.md
├── cmd
│   └── at-mirror
│       ├── cmd
│       │   ├── acct-analyze.go
│       │   ├── acct-expand.go
│       │   ├── acct-feed.go
│       │   ├── acct-stats.go
│       │   ├── acct-sync.go
│       │   ├── acct.go
│       │   ├── ai-chat.go
│       │   ├── ai-complete.go
│       │   ├── ai-embed.go
│       │   ├── ai-explain.go
│       │   ├── ai-hack.go
│       │   ├── ai-reply.go
│       │   ├── ai-safety.go
│       │   ├── ai-summarize.go
│       │   ├── ai-topics.go
│       │   ├── ai.go
│       │   ├── backfill-describe-repo.go
│       │   ├── backfill-pds-accounts.go
│       │   ├── backfill-plc-logs.go
│       │   ├── backfill-repo-sync.go
│       │   ├── config.go
│       │   ├── db-clear.go
│       │   ├── db-migrate.go
│       │   ├── db-reset.go
│       │   ├── plc-annotate.go
│       │   ├── repo-duckdb.go
│       │   ├── repo-hack.go
│       │   ├── repo-inspect.go
│       │   ├── repo-ls.go
│       │   ├── repo-mst.go
│       │   ├── repo-sqlite.go
│       │   ├── repo-sync.go
│       │   ├── repo-unpack.go
│       │   ├── repo-utils.go
│       │   ├── repo.go
│       │   ├── root.go
│       │   └── run.go
│       └── main.go
├── docker-compose.yml
├── docs
│   └── sources.md
├── dsci
│   └── plc
│       ├── Makefile
│       ├── README.md
│       ├── plc-stats.ipynb
│       └── pyproject.toml
├── env-example
├── go.mod
├── go.sum
├── notes.md
├── pkg
│   ├── acct
│   │   ├── acct.go
│   │   ├── analyze.go
│   │   ├── expand.go
│   │   ├── feed.go
│   │   ├── stats.go
│   │   └── sync.go
│   ├── ai
│   │   ├── ai.go
│   │   ├── chat.go
│   │   ├── complete.go
│   │   ├── embed.go
│   │   ├── explain.go
│   │   ├── hack.go
│   │   ├── input.go
│   │   ├── llamacpp
│   │   │   ├── client.go
│   │   │   └── structs.go
│   │   ├── ollama
│   │   │   ├── client.go
│   │   │   └── structs.go
│   │   ├── reply.go
│   │   ├── safety.go
│   │   ├── summarize.go
│   │   └── topics.go
│   ├── config
│   │   ├── config.go
│   │   └── logging.go
│   ├── db
│   │   ├── client.go
│   │   ├── helpers.go
│   │   ├── models.go
│   │   └── views.go
│   ├── plc
│   │   ├── cbor_gen.go
│   │   ├── did_doc.go
│   │   ├── gen
│   │   │   └── main.go
│   │   └── structs.go
│   ├── repo
│   │   ├── blob.go
│   │   ├── car.go
│   │   ├── duckdb.go
│   │   └── sqlite.go
│   ├── runtime
│   │   ├── backfill-describe-repo.go
│   │   ├── backfill-pds-accounts.go
│   │   ├── backfill-repo-sync.go
│   │   ├── const.go
│   │   ├── metrics.go
│   │   ├── plc.go
│   │   ├── queries.go
│   │   ├── repos.go
│   │   ├── runtime.go
│   │   └── utils.go
│   ├── server
│   │   ├── metrics.go
│   │   └── server.go
│   └── util
│       ├── fix
│       │   ├── postgres_json.go
│       │   └── postgres_json_test.go
│       └── gormzerolog
│           └── logger.go
├── pyproject.toml
└── uv.lock

22 directories, 104 files
```
