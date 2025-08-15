
-- extract likes
INSERT INTO refs (source, did, nsid, rkey)
SELECT
  r.cuid as source,
  split_part(json_extract_string(r.record, '$.subject.uri'), '/', 3) as did,
  split_part(json_extract_string(r.record, '$.subject.uri'), '/', 4) as nsid,
  split_part(json_extract_string(r.record, '$.subject.uri'), '/', 5) as rkey
FROM
  records AS r
WHERE
  r.nsid = 'app.bsky.feed.like'
  AND r.cuid NOT IN (SELECT source FROM refs);
