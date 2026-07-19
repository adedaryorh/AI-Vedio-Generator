# AI Video Generator

This project is an MVP for an AI-powered system that collects stories, generates videos, and posts them to multiple social media platforms (Instagram, YouTube, TikTok) automatically. It consists of Python and Go microservices, a PostgreSQL database, Redis, RabbitMQ for message queuing, and Docker-based orchestration.

## IMPLEMENTATION STATUS: ✅ CORE FUNCTIONALITY IMPLEMENTED WITH RABBITMQ INTEGRATION

All services now have working implementations beyond the initial placeholder code, including RabbitMQ integration for improved service decoupling and reliability.

### ✅ **Story Collector Service**
- Fetches stories from multiple sources (Project Gutenberg, Islamic stories, African folklore)
- Enhances stories using OpenAI GPT-3.5-Turbo for better video narration
- Stores stories in PostgreSQL with metadata
- REST API endpoints for story management

### ✅ **Video Generator Service** 
- Converts stories into videos using AI-generated narration (ElevenLabs TTS)
- Creates background visuals from Unsplash based on story content
- Uses MoviePy for video assembly and editing
- Outputs vertical videos optimized for social media (1080x1920)
- REST API for video generation and status tracking
- processes video generation requests from the queue**
- **Publishes "video.ready_for_posting" messages when videos are ready**

### ✅ **Content Manager Service**
- Manages video processing queue (pending → processing → ready)
- Tracks video metadata and storage paths
- Provides statistics on Story/Video processing
- **Publishes "video.ready_for_processing" messages to RabbitMQ when videos are queued**
- **Consumes "video.ready_for_posting" messages to update video status**
- REST API for video and queue management

### ✅ **Social Media Bot Service**
- Supports posting to multiple platforms: Instagram, YouTube, TikTok
- Generates relevant captions and hashtags based on content
- Supports scheduled posts and immediate publishing
- Provides engagement analytics and reporting
- **Consumes "video.ready_for_posting" messages to automatically trigger social media posting**
- REST API for posting and analytics

## 🏗️ **ARCHITECTURE OVERVIEW**

```
story-video-bot/
├── python-services/
│   ├── ai-pipeline/         # Combined story collection, enhancement & video generation
│   ├── story-collector/     # Legacy standalone implementation
│   ├── video-generator/     # Legacy standalone implementation
│   └── shared/              # Shared Python utilities
├── go-services/
│   ├── content-manager/     # Video queue, storage, social publishing & analytics
│   ├── social-media-bot/    # Legacy standalone implementation (not used by Compose)
│   └── shared/              # Shared Go utilities
├── database/
│   └── migrations/          # PostgreSQL schema
├── docker/                  # Docker Compose orchestration
├── config/                  # Environment configuration
└── scripts/                 # Utility scripts (seeding, etc.)
```

## 🚀 **QUICK START**

### 1. **Setup Environment**
```bash
# Copy example env files and configure API keys
cp config/*.example config/*.env
# Edit .env files to add your API keys:
# - OPENAI_API_KEY (for story enhancement)
# - ELEVENLABS_API_KEY (for text-to-speech)
# - UNSPLASH_ACCESS_KEY (for background images)
# - Instagram API credentials (for Instagram posting)
# - YouTube API credentials (for YouTube upload)
# - TikTok API credentials (for TikTok upload)
# - RABBITMQ_USER and RABBITMQ_PASS (for RabbitMQ authentication)
```

### 2. **Start Services**
From the repository root:

```bash
docker compose up --build
```

### 3. **Initialize Database**
```bash
# In another terminal, while containers are running:
docker compose exec ai-pipeline alembic upgrade head
docker compose exec ai-pipeline python scripts/seed_stories.py
```

### 4. **Access Services**
- AI Story + Video Pipeline: `http://localhost:8001/docs`
- Content Manager + Social Publisher: `http://localhost:9001/health`
- API Documentation:
  - Combined Python API: `http://localhost:8001/docs`

## 🔧 **WORKFLOW WITH RABBITMQ**

1. **Story Collection**: Use `/collect` endpoint to fetch stories from various sources
2. **Story Enhancement**: Use `/enhance/{story_id}` to improve stories with AI
3. **Video Generation**: Submit story ID to `/generate` to create video
4. **Content Management**: 
   - Content Manager automatically queues videos (pending status)
   - Publishes "video.ready_for_processing" to RabbitMQ
   - Video Generator consumes processing requests
   - Video Generator publishes "video.ready_for_posting" when done
   - Content Manager updates video status to "ready"
   - Social Media Bot consumes ready messages and auto-posts
   - OR use manual posting via `/post` endpoint
5. **Social Media Publishing**: 
   - Automatic via RabbitMQ messaging
   - Manual via `/post` endpoint for immediate posting
   - `/schedule` endpoint for scheduled posting

## 📚 **API DOCUMENTATION**

Each service provides interactive API documentation:
- **Story Collector**: Swagger UI at `/docs` endpoint
- **Video Generator**: Swagger UI at `/docs` endpoint
- **Content Manager & Social Media Bot**: REST endpoints accessible via HTTP

## ⚙️ **CONFIGURATION**

### Required Environment Variables:
- `DATABASE_URL`: PostgreSQL connection string
- `OPENAI_API_KEY`: For story enhancement (OpenAI)
- `ELEVENLABS_API_KEY`: For text-to-speech narration  
- `UNSPLASH_ACCESS_KEY`: For background image sourcing
- `RABBITMQ_USER`: RabbitMQ username (default: guest)
- `RABBITMQ_PASS`: RabbitMQ password (default: guest)
- Social Media API credentials (for production deployment):
  - Instagram: `INSTAGRAM_ACCESS_TOKEN`, `INSTAGRAM_ACCOUNT_ID`
  - YouTube: `YOUTUBE_API_KEY`, `YOUTUBE_CHANNEL_ID`
  - TikTok: `TIKTOK_ACCESS_KEY`, `TIKTOK_SECRET_KEY`

### Optional:
- `REDIS_URL`: For Celery task queue (if extending with async processing)

## 🛠️ **DEVELOPMENT**

### Testing
```bash
# Python services
cd python-services/*/ && pytest

# Go services  
cd go-services/*/ && go test ./...
```

### Database Management
```bash
# Apply migrations (run while ai-pipeline is running)
docker compose exec ai-pipeline alembic upgrade head

# Seed initial data
docker compose exec ai-pipeline python scripts/seed_stories.py
```

### Viewing Logs
```bash
# View logs for all services
docker compose logs -f

# View logs for specific service
docker compose logs -f ai-pipeline
```

## 📝 **IMPLEMENTATION NOTES**

### Story Collector
- Sources: Project Gutenberg, Islamic Stories API, African Folklore API (mock data for MVP)
- AI Enhancement: Uses OpenAI GPT-3.5-Turbo to make stories more engaging for video
- Processing: Marks stories as processed after enhancement

### Video Generator
- Narration: ElevenLabs TTS for natural-sounding speech
- Visuals: Unsplash API for relevant background images
- Assembly: MoviePy for combining audio, images, and text overlays
- Format: MP4, 1080x1920 (optimized for Instagram Reels, YouTube Shorts, TikTok)
- **Messaging**: Consumes `video.process` from RabbitMQ and publishes `video.ready`

### Content Manager + Social Publisher
- Queue Management: Tracks video processing status (pending → processing → ready)
- Storage: Tracks file paths and metadata in PostgreSQL
- Social Publishing: Provides `/post`, `/schedule`, `/posts`, `/analytics`, and `/platforms`
- **Messaging**: Publishes `video.process` and consumes `video.ready`
- Statistics: Provides counts of stories, videos, and processing states

### Social publishing inside Content Manager
- Platform Support: Instagram, YouTube, TikTok (simulation mode ready for real API)
- Smart Hashtags: Generates relevant hashtags based on content analysis
- Scheduling: Supports both immediate and scheduled posting
- **Messaging**: Consumes `video.ready` to auto-trigger posting
- Analytics: Tracks engagement metrics (likes, comments, shares, views, saves)

## 🔜 **FUTURE ENHANCEMENTS**

1. **Real Social Media API Integration**: Replace simulations with actual platform APIs
2. **Enhanced Error Handling**: Implement retry mechanisms, dead letter queues for failed messages
3. **Monitoring**: Add Prometheus/Grafana for metrics and logging
4. **CD/CI**: Automated testing and deployment pipelines
5. **Additional Story Sources**: Add more cultural story repositories
6. **Video Templates**: Multiple video styles and aspect ratios for different platforms
7. **Advanced AI Features**: 
   - Automatic scene detection and matching
   - Custom voice training for brand consistency
   - A/B testing for thumbnails and captions
8. **Platform-Specific Optimization**: Tailor videos for each platform's requirements
9. **Rate Limiting**: Implement API rate limiting for social media platforms
10. **Caching**: Add Redis caching for frequently accessed data

## 📄 **LICENSE**

MIT License - see LICENSE file for details.# AI-Vedio-Generator
