-- SQL for indexing likes
CREATE TABLE IF NOT EXISTS likes (
  uri TEXT PRIMARY KEY,
  cid TEXT NOT NULL,
  subject_uri TEXT NOT NULL,
  indexed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
