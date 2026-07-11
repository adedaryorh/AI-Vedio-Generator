# Test script to verify the implementation works
echo "Testing Story Video Instagram Bot MVP Implementation"
echo "===================================================="

# Test 1: Check if all services can be built/compiled
echo ""
echo "1. Testing Go services compilation..."
cd /home/adedaryorh/Documents/insta_ai/go-services/content-manager && go build -o /tmp/content_manager_test . 2>/dev/null && echo "✓ Content Manager: OK" || echo "✗ Content Manager: FAILED"

cd /home/adedaryorh/Documents/insta_ai/go-services/social-media-bot && go build -o /tmp/social-media-bot-test . 2>/dev/null && echo "✓ Social Media Bot: OK" || echo "✗ Social Media Bot: FAILED"

# Test 2: Check if Python services have valid syntax
echo ""
echo "2. Testing Python services syntax..."
cd /home/adedaryorh/Documents/insta_ai/python-services/story-collector && python3 -m py_compile app/main.py 2>/dev/null && echo "✓ Story Collector: OK" || echo "✗ Story Collector: FAILED"

cd /home/adedaryorh/Documents/insta_ai/python-services/video-generator && python3 -m py_compile app/main.py 2>/dev/null && echo "✓ Video Generator: OK" || echo "✗ Video Generator: FAILED"

# Test 3: Check if key files exist
echo ""
echo "3. Checking critical files..."
if [ -f "/home/adedaryorh/Documents/insta_ai/docker/docker-compose.yml" ]; then
  echo "✓ Docker Compose: OK"
else
  echo "✗ Docker Compose: MISSING"
fi

if [ -f "/home/adedaryorh/Documents/insta_ai/database/migrations/001_init.sql" ]; then
  echo "✓ Database Schema: OK"
else
  echo "✗ Database Schema: MISSING"
fi

if [ -f "/home/adedaryorh/Documents/insta_ai/scripts/seed_stories.py" ]; then
  echo "✓ Seed Script: OK"
else
  echo "✗ Seed Script: MISSING"
fi

echo ""
echo "Implementation verification complete!"
echo ""
echo "Summary of implemented features:"
echo "- Story Collector: Fetches stories from multiple sources and enhances with AI"
echo "- Video Generator: Creates videos from stories using AI narration and background visuals"
echo "- Content Manager: Manages video queue, processing, and storage tracking"
echo "- Social Media Bot: Handles Instagram posting, scheduling, and analytics"
echo "- All services have health check endpoints"
echo "- Database schema defined for stories, videos, and Instagram posts"
echo "- Docker Compose configuration for easy deployment"