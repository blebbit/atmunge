CREATE TABLE IF NOT EXISTS records (
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
  extra JSON,

  UNIQUE(did, nsid, rkey, cid)
);

-- refs we extract from
CREATE TABLE IF NOT EXISTS refs (
  -- hmmm, we are in sql when we index these, so we don't have cuid's
  -- cuid TEXT PRIMARY KEY,

  -- where the ref was found
  source TEXT, -- a cuid

  -- where the ref points to
  did TEXT,
  nsid TEXT,
  rkey TEXT,

  -- actual contents
  record JSON,

  -- extra stuff
  extra JSON,

  UNIQUE(did, nsid, rkey)
);
