# Content Manager Service

Manages video queue, metadata, and file storage for the story video bot.

## Features
- Video queue management
- Metadata storage (PostgreSQL)
- S3-compatible file storage
- REST API (Gin)
- Quality control and scheduling

## Setup
1. Copy `../../config/content-manager.env.example` to `.env`
2. Install Go 1.21+
3. Run: `go run main.go`

## API
- Runs on port 9001

## Environment Variables
See `.env.example` for required keys. 