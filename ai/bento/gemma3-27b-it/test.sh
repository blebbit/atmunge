#!/bin/bash
set -euo pipefail

time curl -X 'POST' \
  'http://localhost:3000/chat' \
  -H 'Accept: application/json' \
  -H 'Content-Type: application/json' \
  -d @example.json