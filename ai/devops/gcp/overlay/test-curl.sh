#!/bin/bash
set -eou pipefail

time curl -X 'POST' \
  'http://localhost:3000/check' \
  -H 'Accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{}'