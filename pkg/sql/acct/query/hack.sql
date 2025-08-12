INSERT INTO refs (a_nsid, a_rkey, did, nsid, rkey)
SELECT
  nsid as a_nsid,
  rkey as a_rkey,
  split_part(record->>'$.subject.uri', '/', 3) as did,
  split_part(record->>'$.subject.uri', '/', 4) as nsid,
  split_part(record->>'$.subject.uri', '/', 5) as rkey
FROM records
  WHERE nsid = 'app.bsky.feed.like';
