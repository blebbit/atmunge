SELECT
  nsid, count(*) as total
FROM
  records
GROUP BY nsid
ORDER BY total desc;