# Content Manager + Social Publisher

This is the single Go backend for video queue/storage management and social publishing.

Social capabilities include Redis caching, request rate limiting, Prometheus metrics,
platform circuit breakers and retries, RabbitMQ automatic publishing, scheduling,
post cancellation/retry, analytics, and simulated or credential-backed platform clients.

Manages video queue, metadata, and file storage for the story video bot.

## Features
- Video queue management
- Metadata storage (PostgreSQL)
- S3-compatible file storage
- REST API (Gin)
- Quality control and scheduling

## Setup
1. Copy `.env.local.example` to `.env.local`
2. Add any optional platform credentials to `.env.local`
3. Install Go 1.25+
4. Run: `go run .`

## API
- Runs on port 9001

## Environment Variables
See `.env.local.example` for local-development keys.
