# AI Pipeline

This is the single Python backend for story collection, AI enhancement, and video generation.

## Responsibilities

- Lists and collects stories from the existing Gutenberg, Islamic, and African sources.
- Enhances raw stories with OpenAI.
- Generates vertical MP4 videos with narration, background visuals, and title overlays.
- Exposes the original REST routes on one port (`8001`).
- Consumes durable `video.process` jobs from the `video_events` RabbitMQ exchange.
- Updates the existing video row through `pending`, `processing`, `ready`, or `failed`.
- Publishes `video.ready` after successful generation.

## Run locally

From the repository root:

```bash
docker compose up -d postgres redis rabbitmq ai-pipeline
```

Check the service:

```bash
curl http://localhost:8001/health
```

Useful endpoints:

- `GET /stories`
- `POST /collect?source=all`
- `POST /enhance/{story_id}`
- `POST /generate?story_id={story_id}`
- `GET /status/{video_id}`
- `GET /health`

For local development, copy `.env.local.example` to `.env.local`. The checked-in
example uses `localhost` for PostgreSQL, Redis, and RabbitMQ. Docker Compose injects
the container `.env` separately.

## Database migrations

Alembic migrations live in `alembic/` and run automatically before the API starts.
To run them manually inside the container:

```bash
alembic upgrade head
```
