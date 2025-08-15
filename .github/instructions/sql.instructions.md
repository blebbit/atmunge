---
applyTo: "**/*.sql"
---

We have two databases

1. (app) is a postgres database managed with Gorm
2. (acct) is a duckdb database managed with raw SQL

We have the following SQL files

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
          count_nsid.sql
          count_records.sql
          hack.sql
          refs-test.sql