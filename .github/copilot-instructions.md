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

```
.
├── cmd # the core command structure
│   └── at-mirror
│       ├── cmd
│       └── main.go
├── docker-compose.yml
├── Dockerfile
├── docs
│   └── sources.md
├── env-example
├── go.mod
├── go.sum
├── LICENSE
├── Makefile
├── notes.md
├── pkg # the reusable code, functions, and packages
│   ├── config
│   │   ├── config.go
│   │   └── logging.go
│   ├── db
│   │   ├── client.go
│   │   ├── helpers.go
│   │   ├── models.go
│   │   └── views.go
│   ├── plc
│   │   ├── cbor_gen.go
│   │   ├── did_doc.go
│   │   ├── gen
│   │   └── structs.go
│   ├── repo
│   │   ├── car.go
│   │   └── sqlite.go
│   ├── runtime # holds most of the core business logic
│   │   ├── backfill-describe-repo.go
│   │   ├── backfill-get-repo.go
│   │   ├── backfill-pds-accounts.go
│   │   ├── backfill-repo-sync.go
│   │   ├── const.go
│   │   ├── metrics.go
│   │   ├── plc.go
│   │   ├── queries.go
│   │   ├── repos.go
│   │   ├── runtime.go
│   │   └── utils.go
│   ├── server
│   │   ├── metrics.go
│   │   └── server.go
│   └── util
│       ├── fix
│       └── gormzerolog
└── README.md
```
