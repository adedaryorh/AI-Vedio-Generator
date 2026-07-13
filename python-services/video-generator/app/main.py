from fastapi import FastAPI, Query, BackgroundTasks, HTTPException
from typing import List, Optional, Dict, Any
import os
import logging
import json
import asyncio
from datetime import datetime
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
from sqlalchemy import text
import openai
import requests
from moviepy.editor import *
from moviepy.config import change_settings
import tempfile
import urllib.request
from PIL import Image, ImageDraw, ImageFont
import numpy as np

load_dotenv()

app = FastAPI(title="Video Generator Service")

# Database setup
DATABASE_URL = os.getenv("DATABASE_URL", "postgresql+asyncpg://storybot:storybotpass@postgres:5432/storybotdb")
engine = create_async_engine(DATABASE_URL)
AsyncSessionLocal = sessionmaker(engine, class_=AsyncSession, expire_on_commit=False)

# AI Services
openai.api_key = os.getenv("OPENAI_API_KEY")
ELEVENLABS_API_KEY = os.getenv("ELEVENLABS_API_KEY")
UNSPLASH_ACCESS_KEY = os.getenv("UNSPLASH_ACCESS_KEY")

# Video settings
VIDEO_WIDTH = 1080
VIDEO_HEIGHT = 1920  # Instagram Reels/Stories format
FPS = 30
TEMP_DIR = tempfile.mkdtemp()

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@app.get("/health")
async def health():
    return {"status": "ok"}

@app.post("/generate")
async def generate_video(story_id: int = Query(...), background_tasks: BackgroundTasks = None):
    background_tasks.add_task(generate_video_for_story, story_id)
    return {"message": f"Started generating video for story {story_id}"}

@app.get("/status/{video_id}")
async def video_status(video_id: int):
    async with AsyncSessionLocal() as session:
        result = await session.execute(
            text("SELECT id, status, duration, file_path FROM videos WHERE id = :id"),
            {"id": video_id}
        )
        video = result.fetchone()
        if not video:
            raise HTTPException(status_code=404, detail="Video not found")

        return {
            "video_id": video[0],
            "status": video[1],
            "duration": video[2],
            "file_path": video[3]
        }

async def generate_video_for_story(story_id: int):
    try:
        async with AsyncSessionLocal() as session:
            # Get the story
            result = await session.execute(
                text("SELECT id, title, content FROM stories WHERE id = :id AND processed = true"),
                {"id": story_id}
            )
            story = result.fetchone()
            if not story:
                logger.warning(f"Story {story_id} not found or not processed")
                return

            story_id, title, content = story

            logger.info(f"Generating video for story {story_id}: {title}")

            # 1. Create video segments
            video_clips = []

            # 2. Generate narration audio
            narration_audio = await generate_narration_audio(content, story_id)
            if narration_audio:
                video_clips.append(narration_audio)

            # 3. Get background visuals
            background_clips = await get_background_visuals(content, narration_audio.duration if narration_audio else 10)
            video_clips.extend(background_clips)

            # 4. Add title overlay
            title_clip = create_title_clip(title, narration_audio.duration if narration_audio else 10)
            if title_clip:
                video_clips.append(title_clip)

            # 5. Composite video
            if video_clips:
                final_video = CompositeVideoClip(video_clips, size=(VIDEO_WIDTH, VIDEO_HEIGHT))

                # Set duration
                if narration_audio:
                    final_video = final_video.set_duration(narration_audio.duration)
                else:
                    final_video = final_video.set_duration(10)  # default duration

                # Write video
                output_filename = f"/app/videos/story_{story_id}_{int(datetime.now().timestamp())}.mp4"
                os.makedirs(os.path.dirname(output_filename), exist_ok=True)

                final_video.write_videofile(
                    output_filename,
                    fps=FPS,
                    codec='libx264',
                    audio_codec='aac',
                    temp_audiofile='/temp-audio.m4a',
                    remove_temp=True,
                    verbose=False,
                    logger=None
                )

                # Insert video record
                await session.execute(
                    text("""
                        INSERT INTO videos (story_id, file_path, duration, format, resolution, status, metadata)
                        VALUES (:story_id, :file_path, :duration, :format, :resolution, :status, :metadata)
                    """),
                    {
                        "story_id": story_id,
                        "file_path": output_filename,
                        "duration": int(final_video.duration),
                        "format": "mp4",
                        "resolution": f"{VIDEO_WIDTH}x{VIDEO_HEIGHT}",
                        "status": "ready",
                        "metadata": json.dumps({
                            "title": title,
                            "generated_at": datetime.now().isoformat(),
                            "has_narration": narration_audio is not None
                        })
                    }
                )
                await session.commit()

                logger.info(f"Video generated successfully: {output_filename}")
            else:
                logger.error(f"No video clips generated for story {story_id}")
                # Insert failed video record
                await session.execute(
                    text("""
                        INSERT INTO videos (story_id, file_path, duration, format, resolution, status, metadata)
                        VALUES (:story_id, :file_path, :duration, :format, :resolution, :status, :metadata)
                    """),
                    {
                        "story_id": story_id,
                        "file_path": "",  # No file generated
                        "duration": 0,
                        "format": "",
                        "resolution": "",
                        "status": "failed",
                        "metadata": json.dumps({
                            "title": title,
                            "error": "No video clips generated",
                            "generated_at": datetime.now().isoformat()
                        })
                    }
                )
                await session.commit()

    except Exception as e:
        logger.error(f"Error generating video for story {story_id}: {e}")
        # Insert failed video record
        try:
            async with AsyncSessionLocal() as session:
                await session.execute(
                    text("""
                        INSERT INTO videos (story_id, file_path, duration, format, resolution, status, metadata)
                        VALUES (:story_id, :file_path, :duration, :format, :resolution, :status, :metadata)
                    """),
                    {
                        "story_id": story_id,
                        "file_path": "",  # No file generated due to error
                        "duration": 0,
                        "format": "",
                        "resolution": "",
                        "status": "failed",
                        "metadata": json.dumps({
                            "title": title if 'title' in locals() else "Unknown",
                            "error": "Video generation failed due to exception",
                            "generated_at": datetime.now().isoformat() if 'datetime' in locals() else "",
                            "exception": str(e) if 'e' in locals() else "Unknown error"
                        })
                    }
                )
                await session.commit()
        except:
            pass  # If we can't even log the failure, there's nothing more we can do

async def generate_narration_audio(text: str, story_id: int):
    try:
        # Use ElevenLabs for TTS
        if ELEVENLABS_API_KEY:
            url = "https://api.elevenlabs.io/v1/text-to-speech/21m00Tcm4TlvDq8ikWAM"

            headers = {
                "Accept": "audio/mpeg",
                "Content-Type": "application/json",
                "xi-api-key": ELEVENLABS_API_KEY
            }

            data = {
                "text": text,
                "model_id": "eleven_monolingual_v1",
                "voice_settings": {
                    "stability": 0.5,
                    "similarity_boost": 0.5
                }
            }

            response = requests.post(url, json=data, headers=headers)
            if response.status_code == 200:
                audio_path = f"{TEMP_DIR}/narration_{story_id}.mp3"
                with open(audio_path, "wb") as f:
                    f.write(response.content)

                audio_clip = AudioFileClip(audio_path)
                return audio_clip

        # Fallback to silence
        logger.warning("ElevenLabs not configured, using silence")
        from moviepy.editor import AudioClip
        silence = AudioClip(lambda t: 0, duration=5)  # 5 seconds of silence
        return silence

    except Exception as e:
        logger.error(f"Error generating narration: {e}")
        return None

async def get_background_visuals(text: str, duration: float):
    clips = []
    try:
        # Get images from Unsplash based on keywords in the story
        keywords = extract_keywords(text)

        # Calculate how long each image should show
        num_images = min(len(keywords), 3)  # Max 3 images
        if num_images == 0:
            num_images = 1
            keywords = ["nature"]  # fallback

        duration_per_image = duration / num_images if num_images > 0 else duration

        for i, keyword in enumerate(keywords[:num_images]):
            try:
                # Get image from Unsplash
                image_url = get_unsplash_image(keyword)
                if image_url:
                    # Download image
                    image_path = f"{TEMP_DIR}/bg_{story_id}_{i}.jpg"
                    urllib.request.urlretrieve(image_url, image_path)

                    # Create image clip
                    img_clip = ImageClip(image_path).set_duration(duration_per_image)

                    # Resize to fit video dimensions while maintaining aspect ratio
                    img_clip = img_clip.resize(height=VIDEO_HEIGHT).set_pos('center')
                    if img_clip.w > VIDEO_WIDTH:
                        img_clip = img_clip.resize(width=VIDEO_WIDTH).set_pos('center')

                    clips.append(img_clip)

            except Exception as e:
                logger.error(f"Error processing background image for keyword {keyword}: {e}")
                # Create a colored background as fallback
                color_clip = ColorClip(size=(VIDEO_WIDTH, VIDEO_HEIGHT), color=(64, 64, 64), duration=duration_per_image)
                clips.append(color_clip)

        # If we didn't get enough images, fill with solid color
        while len(clips) < num_images:
            color_clip = ColorClip(size=(VIDEO_WIDTH, VIDEO_HEIGHT), color=(32, 32, 32), duration=duration_per_image)
            clips.append(color_clip)

    except Exception as e:
        logger.error(f"Error getting background visuals: {e}")
        # Fallback: solid background
        clips.append(ColorClip(size=(VIDEO_WIDTH, VIDEO_HEIGHT), color=(32, 32, 32), duration=duration))

    return clips

def extract_keywords(text: str):
    # Simple keyword extraction - in production would use NLP
    words = text.lower().split()
    # Filter out common words and get meaningful ones
    stop_words = {'the', 'a', 'an', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for', 'of', 'with', 'by', 'is', 'are', 'was', 'were', 'be', 'been', 'being', 'have', 'has', 'had', 'do', 'does', 'did', 'will', 'would', 'could', 'should', 'may', 'might', 'must', 'shall', 'can'}
    keywords = [word.strip('.,!?;:"') for word in words if word not in stop_words and len(word) > 3]
    # Return unique keywords, limited
    return list(dict.fromkeys(keywords))[:5]

def get_unsplash_image(keyword: str):
    try:
        if not UNSPLASH_ACCESS_KEY:
            # Return a placeholder or use a different method
            return f"https://source.unsplash.com/random/{VIDEO_WIDTH}x{VIDEO_HEIGHT}?{keyword}"

        url = "https://api.unsplash.com/photos/random"
        params = {
            "query": keyword,
            "orientation": "portrait",
            "width": VIDEO_WIDTH,
            "height": VIDEO_HEIGHT
        }
        headers = {
            "Authorization": f"Client-ID {UNSPLASH_ACCESS_KEY}"
        }

        response = requests.get(url, params=params, headers=headers)
        if response.status_code == 200:
            data = response.json()
            return data["urls"]["regular"]
        else:
            logger.warning(f"Unsplash API error: {response.status_code}")
            return f"https://source.unsplash.com/random/{VIDEO_WIDTH}x{VIDEO_HEIGHT}?{keyword}"
    except Exception as e:
        logger.error(f"Error fetching Unsplash image: {e}")
        return f"https://source.unsplash.com/random/{VIDEO_WIDTH}x{VIDEO_HEIGHT}?{keyword}"

def create_title_clip(title: str, duration: float):
    try:
        # Create a text clip for the title
        txt_clip = TextClip(
            title,
            fontsize=70,
            color='white',
            font='Arial-Bold',
            stroke_color='black',
            stroke_width=2
        )
        txt_clip = txt_clip.set_duration(duration)
        txt_clip = txt_clip.set_pos(('center', 'top'))

        # Add semi-transparent background for better readability
        bg_clip = ColorClip(size=(VIDEO_WIDTH, 100), color=(0, 0, 0), duration=duration)
        bg_clip = bg_clip.set_opacity(0.5)
        bg_clip = bg_clip.set_pos(('center', 'top'))

        return CompositeVideoClip([bg_clip, txt_clip])
    except Exception as e:
        logger.error(f"Error creating title clip: {e}")
        return None

def load_dotenv():
    from dotenv import load_dotenv
    load_dotenv()

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8002)