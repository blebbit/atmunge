# PLC mirror

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

You can directly replace `https://plc.directory` with a URL to the exposed port
(11004 by default).

Note that on the first run it will take quite a few hours to download everything,
and the mirror with respond with 500 if it's not caught up yet.

As of March 2025, the DB is around:

- 45G in postgres
- 7G as a backup

... record count