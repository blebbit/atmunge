# plc-utils
ATProto PLC Utilities

Plan:

- plc backup & snapshot
- plc mirror & syncing
- extras that are helpful to apps building on atproto
  - provide handle based lookup
  - account collection list
  - denormalized did doc details for faster queries
  - `did:web:...` support 
  - clients which fall back to official PLC (or other mirrors for read only)
    - proxy write requests?
  - authnz setup
- provide public read-only mirror with generous rate limits
- provide setups for multiple clouds
- fun data science & graphs
- public stats dashboard and explorer




### References & Related

- https://web.plc.directory/
- https://github.com/bsky-watch/plc-mirror 



### PLC error groups


##### first event, https://uwu, 92000 -> 172000

```
atmunge=# SELECT filtered, COUNT(*) FROM plc_log_entries WHERE id < 1000000 GROUP BY filtered;
 filtered | count  
----------+--------
        0 | 988248
        1 |     24
        3 |  11727
(3 rows)

atmunge=# SELECT notes, COUNT(*) FROM plc_log_entries WHERE id < 1000000 GROUP BY notes;
                   notes                   | count  
-------------------------------------------+--------
 HDL:0:data-x; PDS:known-bad; DOC:make-doc |  11704
 HDL:0:length; PDS:known-bad; DOC:make-doc |     23
 HDL:0:regex                               |      5
 PDS:not-set                               |     17
 PDS:parse                                 |      2
                                           | 988248
(6 rows)
```

