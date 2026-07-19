import os
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.stories import router as stories_router
from app.videos import router as videos_router
from app.messaging import video_consumer


@asynccontextmanager
async def lifespan(_: FastAPI):
    video_consumer.start()
    yield
    video_consumer.stop()

app = FastAPI(
    title="AI Story and Video Pipeline",
    description="Collects and enhances stories, then generates social-ready videos.",
    version="1.0.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=[os.getenv("FRONTEND_URL", "http://localhost:3000")],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(stories_router)
app.include_router(videos_router)


@app.get("/health", tags=["system"])
async def health():
    return {
        "status": "ok",
        "service": "ai-pipeline",
        "features": ["story-collection", "story-enhancement", "video-generation"],
        "rabbitmq_consumer": "connected" if video_consumer.connected else "disconnected",
    }
