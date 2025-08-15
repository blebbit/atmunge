SELECT
  nsid, count(*) as total
FROM
  refs
GROUP BY nsid
ORDER BY total desc;
