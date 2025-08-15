---
applyTo: "**/*.sql"
---

We have two databases

1. (app) is a postgres database managed with Gorm
2. (acct) is a duckdb database managed with raw SQL

You can arbitrary queries on acct with the `acct query` command.

We have the following SQL files.
They are embedded into the Go program in `./pkg/sql/embed.go`

./
  pkg/
    sql/
      acct/
        index/
          0100_extract_refs.sql
        migrations/
          0001_init.sql
        query/
          count_likes_by_did.sql
          count_records.sql
          count_records_by_nsid.sql
          count_refs_by_nsid.sql
          hack.sql
          refs-test.sql