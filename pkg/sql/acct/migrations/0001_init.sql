CREATE TABLE IF NOT EXISTS records (
  -- for relational queries
  cuid TEXT PRIMARY KEY,

  -- time informations
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  indexed_at TIMESTAMP,

  -- where the ref points to
  did TEXT,
  nsid VARCHAR,
  rkey VARCHAR,
  cid VARCHAR,

  -- actual contents
  record JSON,

  -- extra stuff
  extra JSON
);

-- refs we extract from 
CREATE TABLE IF NOT EXISTS refs (
  cuid TEXT PRIMARY KEY,

  -- where the ref was found
  source TEXT -- a cuid

  -- where the ref points to
  did TEXT,
  nsid TEXT,
  rkey TEXT,

  -- actual contents
  record JSON,

  -- extra stuff
  extra JSON
);
