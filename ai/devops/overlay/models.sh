#!/bin/bash
set -euo pipefail


mkdir -p models
pushd models
uv init
uv add \
  accelerate \
  bentoml \
  pillow \
  requests \
  torch \
  torchvision \
  transformers
