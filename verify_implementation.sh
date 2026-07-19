#!/bin/bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "Verifying AI Video Generator"

echo "1. Go Content Manager"
(cd "$ROOT_DIR/go-services/content-manager" && go test ./...)

echo "2. Unified Python AI Pipeline"
(cd "$ROOT_DIR/python-services/ai-pipeline" && \
  PYTHONPYCACHEPREFIX=/tmp/ai-pipeline-pycache python3 -m py_compile \
  app/main.py app/stories.py app/videos.py app/messaging.py \
  alembic/env.py alembic/versions/0001_initial.py)

echo "3. Docker Compose configuration"
(cd "$ROOT_DIR" && docker compose config --quiet)

echo "Verification complete."
