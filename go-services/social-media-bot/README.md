# Social Media Bot Service

Automates social media posting, engagement, and analytics for generated videos across multiple platforms (Instagram, YouTube, TikTok).

## Features Implemented

- **Instagram API integration** (with fallback to simulation)
- **YouTube Data API integration** (with fallback to simulation)
- **TikTok API integration** (with fallback to simulation - note: TikTok API access is restricted)
- Caption/hashtag generation based on video content
- Scheduler and engagement automation simulation
- Analytics tracking
- REST API (net/http)
- RabbitMQ integration for event-driven processing

## Endpoints

- `GET /health` - Health check endpoint with platform API status
- `POST /post` - Post a video to specified social media platform(s)
- `POST /schedule` - Schedule a post for later publishing
- `GET /analytics` - Get engagement analytics across platforms
- `GET /posts` - Get list of posted videos
- `GET /platforms` - Get list of supported social media platforms

## Environment Variables

See `.env.example` for required keys.

### Instagram
- `INSTAGRAM_APP_ID` - Instagram App ID
- `INSTAGRAM_APP_SECRET` - Instagram App Secret
- `INSTAGRAM_ACCESS_TOKEN` - Instagram Access Token

### YouTube
- `YOUTUBE_API_KEY` - YouTube Data API v3 Key

### TikTok (Note: Access is restricted)
- `TIKTOK_CLIENT_KEY` - TikTok Client Key
- `TIKTOK_CLIENT_SECRET` - TikTok Client Secret
- `TIKTOK_ACCESS_TOKEN` - TikTok Access Token

### Infrastructure
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
- `RABBITMQ_USER` - RabbitMQ username
- `RABBITMQ_PASS` - RabbitMQ password
- `RABBITMQ_HOST` - RabbitMQ host
- `RABBITMQ_PORT` - RabbitMQ port
- `PORT` - Server port (defaults to 9002)

## Implementation Details

This service implements actual API integrations for social media platforms with graceful fallback to simulation when credentials are not available:

### Instagram Integration
When credentials are provided, uses the Instagram Graph API to:
1. Upload video content
2. Create media container
3. Publish the post
4. Return the Instagram post ID

### YouTube Integration
When credentials are provided, uses the YouTube Data API v3 to:
1. Upload video file
2. Set title, description, and tags
3. Set privacy status
4. Return the YouTube video ID

### TikTok Integration
Note: TikTok's API for video upload is highly restricted and requires business account approval. 
When credentials are provided and access is granted, would use the TikTok API to upload videos.
Currently falls back to simulation due to API access restrictions.

### Fallback Behavior
If API credentials are not available or API calls fail, the service automatically falls back to simulation mode:
- Simulates network delay
- Generates realistic-looking platform IDs
- Logs simulation activities
- Maintains API compatibility

## API Response Format

All endpoints return JSON responses. Successful posts return:
```json
{
  "message": "Successfully posted to instagram",
  "platform_id": "ig_123456789",
  "post": {
    "id": 1,
    "video_id": 101,
    "platform": "instagram",
    "platform_id": "ig_123456789",
    "caption": "Your caption here",
    "hashtags": ["#example", "#test"],
    "posted_at": "2023-01-01T12:00:00Z",
    "engagement_metrics": {
      "likes": 0,
      "comments": 0,
      "shares": 0,
      "saves": 0,
      "views": 0
    },
    "status": "posted"
  }
}
```

## Setup

1. Copy `../../config/social-media-bot.env.example` to `.env`
2. Fill in required environment variables for the platforms you want to use
3. Install Go 1.21+
4. Run: `go run main.go`

The service will start on port 9002 by default.

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o social-media-bot main.go
```

## Architecture

The service follows a modular design:
- Platform-specific functions are isolated for easy maintenance
- Configuration is loaded at startup
- RabbitMQ consumer runs as a goroutine for async processing
- HTTP handlers are decoupled from business logic
- Error handling includes graceful degradation to simulation mode