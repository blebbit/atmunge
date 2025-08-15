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

-- extract reposts
INSERT INTO refs (source, did, nsid, rkey)
SELECT
  r.cuid as source,
  split_part(json_extract_string(r.record, '$.subject.uri'), '/', 3) as did,
  split_part(json_extract_string(r.record, '$.subject.uri'), '/', 4) as nsid,
  split_part(json_extract_string(r.record, '$.subject.uri'), '/', 5) as rkey
FROM
  records AS r
WHERE
  r.nsid = 'app.bsky.feed.repost'
  AND r.cuid NOT IN (SELECT source FROM refs);

-- extract follows
INSERT INTO refs (source, did)
SELECT
  r.cuid as source,
  json_extract_string(r.record, '$.subject') as did
FROM
  records AS r
WHERE
  r.nsid = 'app.bsky.graph.follow'
  AND r.cuid NOT IN (SELECT source FROM refs);

-- extract replies to posts
INSERT INTO refs (source, did, nsid, rkey)
SELECT
  r.cuid as source,
  split_part(json_extract_string(r.record, '$.reply.parent.uri'), '/', 3) as did,
  split_part(json_extract_string(r.record, '$.reply.parent.uri'), '/', 4) as nsid,
  split_part(json_extract_string(r.record, '$.reply.parent.uri'), '/', 5) as rkey
FROM
  records AS r
WHERE
  r.nsid = 'app.bsky.feed.post'
  AND json_extract_string(r.record, '$.reply.parent.uri') IS NOT NULL
  AND split_part(json_extract_string(r.record, '$.reply.parent.uri'), '/', 4) IN ('app.bsky.feed.post', 'app.bsky.feed.repost')
  AND r.cuid NOT IN (SELECT source FROM refs);

-- extract list items from subject
INSERT INTO refs (source, did)
SELECT
  r.cuid as source,
  json_extract_string(r.record, '$.subject') as did
FROM
  records AS r
WHERE
  r.nsid = 'app.bsky.graph.listitem'
  AND r.cuid NOT IN (SELECT source FROM refs);
