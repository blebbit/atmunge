-- SQL for indexing posts
CREATE TABLE IF NOT EXISTS refs (
  -- where the ref was found
  a_nsid TEXT,
  a_rkey TEXT,

  -- where the ref points to
  did TEXT,
  nsid TEXT,
  rkey TEXT,

  -- actual contents
  record JSON
);
