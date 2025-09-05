#!/bin/bash
set -eou pipefail

pushd moderation

uv run ./llamaguard4.py