-- SQL for indexing posts
CREATE TABLE IF NOT EXISTS posts (
  uri TEXT PRIMARY KEY,
  cid TEXT NOT NULL,
  reply_parent TEXT,
  reply_root TEXT,
  indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
