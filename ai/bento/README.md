# atmunge/bento

Serves up models


## Setup




## Using

In any of the subdirs, run

```sh
cd <model>
uv run bentoml serve
```

Then you can curl, body depending on the model

```sh
time curl -X 'POST' \
  'http://localhost:3000/check' \
  -H 'Accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "message": "...",
    "imageUrl": "..."
  }'
```
