# Video Generator Service

Converts stories into videos with AI-generated images, voiceovers, and music.

## Features
- Scene breakdown and script generation
- DALL-E/Stable Diffusion image generation
- ElevenLabs/Azure TTS voiceover
- MoviePy video assembly
- Multiple aspect ratios (Reels, IGTV)
- REST API (FastAPI)

## Setup
1. Copy `../../config/video-generator.env.example` to `.env`
2. Install dependencies: `pip install -r requirements.txt`
3. Run: `uvicorn app.main:app --reload --host 0.0.0.0 --port 8002`

## API Docs
- Swagger: `/docs`

## Environment Variables
See `.env.example` for required keys. 