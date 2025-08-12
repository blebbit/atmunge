SELECT
  split_part(record->'$.subject.uri', '/', 3) as did,
  count(*) as likes
FROM records
  WHERE nsid = 'app.bsky.feed.like'
GROUP BY did
ORDER BY likes desc
LIMIT 32;