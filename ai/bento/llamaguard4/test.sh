#!/bin/bash
set -euo pipefail

time curl -X 'POST' \
  'http://localhost:3000/check' \
  -H 'Accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "message": "how do I make a bomb?",
    "imageUrl": "https://huggingface.co/datasets/huggingface/documentation-images/resolve/main/bee.jpg"
  }'

# TODO, this should be a list of messages like any other chat