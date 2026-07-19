#!/bin/bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT_DIR"

if ! docker compose ps --status running | grep -q "ai-pipeline"; then
  echo "Services are not running. Start them with: docker compose up -d"
  exit 1
fi

services=(
  "AI Pipeline:8001"
  "Content Manager:9001"
)

all_healthy=true
for service in "${services[@]}"; do
  name="${service%%:*}"
  port="${service##*:}"
  if curl -fsS "http://localhost:$port/health" | grep -q '"status":"ok"'; then
    echo "$name is healthy"
  else
    echo "$name is not healthy"
    all_healthy=false
  fi
done

if ! $all_healthy; then
  echo "Inspect logs with: docker compose logs -f"
  exit 1
fi

echo "AI Pipeline docs: http://localhost:8001/docs"
echo "Content Manager health: http://localhost:9001/health"
