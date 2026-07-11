# Story Collector Service

Collects and enhances stories from various sources for video generation.

## Features
- Fetches stories from Project Gutenberg, Islamic APIs, African folklore
- Cleans and categorizes content
- Enhances stories using OpenAI GPT-4
- REST API (FastAPI)
- Stores stories in PostgreSQL

## Setup
1. Copy `../../config/story-collector.env.example` to `.env`
2. Install dependencies: `pip install -r requirements.txt`
3. Run: `uvicorn app.main:app --reload --host 0.0.0.0 --port 8001`

## API Docs
- Swagger: `/docs`

## Environment Variables
See `.env.example` for required keys. 