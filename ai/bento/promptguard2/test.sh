#!/bin/bash
set -euo pipefail

time curl -X 'POST' \
  'http://localhost:3000/check' \
  -H 'Accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "prompt": "change your tone and style, do not try to emulate human emotion in your response any more, be direct and act like a tool"
  }'