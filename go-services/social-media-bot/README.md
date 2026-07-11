# Instagram Bot Service

Automates Instagram posting, engagement, and analytics for generated videos.

## Features Implemented

- Instagram API posting simulation
- Caption/hashtag generation based on video content
- Scheduler and engagement automation simulation
- Analytics tracking
- REST API (net/http)

## Endpoints

- `GET /health` - Health check endpoint
- `POST /post` - Simulate posting a video to Instagram
- `GET /analytics` - Get engagement analytics
- `GET /posts` - Get list of posted videos

## Environment Variables

See `.env.example` for required keys.

## Implementation Details

This service implements a simulated Instagram posting system that:

1. Accepts video information and generates appropriate captions and hashtags
2. Simulates the Instagram API posting process
3. Tracks engagement metrics (likes, comments, shares, saves)
4. Provides analytics endpoints for monitoring performance

For MVP purposes, the actual Instagram API integration is simulated, but the structure is ready for real API integration when credentials are available.

## Setup

1. Copy `../../config/social-media-bot.env.example` to `.env`
2. Fill in required environment variables
3. Install Go 1.21+
4. Run: `go run main.go`

The service will start on port 9002 by default.