#!/bin/bash
set -euo pipefail

time curl -X 'POST' \
  'http://localhost:3000/promptguard2' \
  -H 'Accept: application/json' \
  -H 'Content-Type: application/json' \

time curl -X 'POST' \
  'http://localhost:3000/shieldgemma' \
  -H 'Accept: application/json' \
  -H 'Content-Type: application/json' \

time curl -X 'POST' \
  'http://localhost:3000/shieldgemma2' \
  -H 'Accept: application/json' \
  -H 'Content-Type: application/json' \

time curl -X 'POST' \
  'http://localhost:3000/llamaguard4' \
  -H 'Accept: application/json' \
  -H 'Content-Type: application/json' \