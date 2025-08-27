# Blebbit AT Mirror

A collection of tools and utilities for backfilling, mirroring, and analyzing the ATProtocol network.

> [!NOTE]
> This repository is undergoing a major refactoring. Please stay tuned as changes come in!


## Install & Setup

```sh
# install the cli
CGO_ENABLED=1 go install ./cmd/atmunge

# setup the env
cp env-example .env

# start the db
docker compose up -d db

# run db migrations
atmunge db migrate
```

Deps:

- docker
- postgresql tools (psql, pg_dump, pg_restore)

Downloads:

TODO, rename R2 dir to atmunge (really nuke and re-upload)

Database dumps to prefill and save time with

- [plc_log_entries-raw-20250706.sql.zst](https://public.blebbit.dev/atmunge/plc_log_entries-raw-20250706.sql.zst)
- [plc_log_entries-filtered-20250706.sql.zst](https://public.blebbit.dev/atmunge/plc_log_entries-filtered-20250706.sql.zst)
- [pds_repos-2025-0714.sql.zst](https://public.blebbit.dev/atmunge/pds_repos-20250714.sql.zst)
- [account_infos-2025-0714.sql.zst](https://public.blebbit.dev/atmunge/account_infos-20250714.sql.zst)


## Backfilling

We recommend starting from the database backups above. It will save you many hours (24h+)
Either way, the following commands will work from empty or restored database tables.

Restore: (DATE should match the file you downloaded and may not be consistent between commands)

```sh
make DATE=YYYYMMDD restore.plc_log_entries.raw
make DATE=YYYYMMDD restore.pds_repos
make DATE=YYYYMMDD restore.account_infos
```

Backfill:

```sh
# backfill the raw PLC logs (~12h when starting from zero)
atmunge backfill plc-logs [--fliter]

# backfill the pds_repos list (~4h)
atmunge backfill pds-accounts

# backfill the accounts_infos table (~20h)
#   describe repo (status + collections)
#   (also writes to the pds_repos table to update status)
atmunge backfill describe-repo
```

Download Repos:

```sh
# WARN, this will chew up disk, you probably need around 10T for the CAR files (Aug'25)
atmunge backfill repo-sync

# fetch CAR for an account
atmunge repo sync verdverm.com

# convert to a database
atmunge repo duckdb ./data/repos/<did>.car
atmunge repo sqlite ./data/repos/<did>.car
```


## Serving

You can serve you backfills as a unified API,
optionally with mirroring enabled so they stay up to date.



---


Syncs PLC operations log into a local table, and allows resolving `did:plc:`
DIDs without putting strain on https://plc.directory and hitting rate limits.
Also syncs key acct info (did, handle, pds) to a second table for light weight queries.
Several extra endpoints are provided for convenience.

```sh
/<did>               # get DID doc
/info/<did|handle>   # bi-directional lookup of key acct info

/ready     # is the mirror up-to-date
/metrics   # for prometheus
```

## Setup

* Decide where do you want to store the data
* Copy `example.env` to `.env` and edit it to your liking.
    * `POSTGRES_PASSWORD` can be anything, it will be used on the first start of
      `postgres` container to initialize the database.
* `make up`

## Usage

### Public Server

We host a public instance of this server at `https://plc.blebbit.dev`.
Currently it has no rate limits, but this may change if there is abuse or excessive use.

### Self Hosting

You can directly replace `https://plc.directory` with a URL to the exposed port
(11004 by default).

Note that on the first run it will take quite a few hours to download everything,
and the mirror with respond with 500 if it's not caught up yet.

### Snapshots

We also provide direct downloads for the `pg_dump` to shorten the backfill time
or if you want to do anything else with the data once it is in Postgresql.

https://public.blebbit.dev/plc/snapshot/plc-20250307.sql.zst

As of early March 2025, the DB is around

Size:

- postgres: 55G
- snapshot: 8G

Records:

- plc rows: 38M
- did rows: 34M



## Notes

consider returning `handle.invalid` when handle does not match doc
- https://docs.bsky.app/docs/api/com-atproto-identity-resolve-identity

Look into index performance and possible removing second table
- https://medium.com/geekculture/postgres-jsonb-usage-and-performance-analysis-cdbd1242a018


Autocomplete needed:

```sql
CREATE EXTENSION pg_trgm;
CREATE INDEX account_infos_handle_gin_trgm_idx  ON account_infos USING gin  (handle gin_trgm_ops);
```