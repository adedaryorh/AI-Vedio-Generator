#!/bin/bash

# Verification script for Story Video Social Media Bot MVP

echo "🔍 Verifying Story Video Social Media Bot MVP deployment..."
echo ""

# Check if docker-compose is running
if ! docker-compose ps | grep -q "Up"; then
  echo "❌ Docker compose services are not running. Please start them with: docker-compose up -d"
  exit 1
fi

echo "✅ Docker compose services are running"

# Check each service's health
services=(
  "story-collector:8001"
  "video-generator:8002"
  "content-manager:9001"
  "social-media-bot:9002"
)

all_healthy=true

for service in "${services[@]}"; do
  service_name=$(echo $service | cut -d':' -f1)
  port=$(echo $service | cut -d':' -f2)

  if curl -s "http://localhost:$port/health" | grep -q '"status":"ok"'; then
    echo "✅ $service_name is healthy"
  else
    echo "❌ $service_name is not responding correctly"
    all_healthy=false
  fi
done

if $all_healthy; then
  echo ""
  echo "🎉 All services are healthy and running!"
  echo ""
  echo "📋 Service URLs:"
  echo "   - Story Collector API: http://localhost:8001/docs"
  echo "   - Video Generator API: http://localhost:8002/docs"
  echo "   - Content Manager: http://localhost:9001/health"
  echo "   - Social Media Bot: http://localhost:9002/health"
  echo ""
  echo "🚀 To start using the system:"
  echo "   1. Collect stories: POST to http://localhost:8001/collect"
  echo "   2. Enhance stories: POST to http://localhost:8001/enhance/{story_id}"
  echo "   3. Generate videos: POST to http://localhost:8002/generate"
  echo "   4. Post to social media: POST to http://localhost:9002/post"
  exit 0
else
  echo ""
  echo "⚠️  Some services are not healthy. Check the logs with: docker-compose logs -f"
  exit 1
fi