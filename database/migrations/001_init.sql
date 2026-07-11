-- 001_init.sql: Initial schema for story-video-bot

CREATE TABLE stories (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    source VARCHAR(100),
    category VARCHAR(50),
    culture VARCHAR(50),
    tags TEXT[],
    created_at TIMESTAMP DEFAULT NOW(),
    processed BOOLEAN DEFAULT FALSE
);

CREATE TABLE videos (
    id SERIAL PRIMARY KEY,
    story_id INTEGER REFERENCES stories(id),
    file_path VARCHAR(500),
    duration INTEGER,
    format VARCHAR(20),
    resolution VARCHAR(20),
    status VARCHAR(50) DEFAULT 'pending',
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE instagram_posts (
    id SERIAL PRIMARY KEY,
    video_id INTEGER REFERENCES videos(id),
    instagram_id VARCHAR(100),
    caption TEXT,
    hashtags TEXT[],
    posted_at TIMESTAMP,
    engagement_metrics JSONB,
    status VARCHAR(50) DEFAULT 'scheduled'
); 