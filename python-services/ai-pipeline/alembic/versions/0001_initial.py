"""Create the story and video pipeline tables.

Revision ID: 0001_initial
Revises:
"""
from typing import Sequence, Union

from alembic import op

revision: str = "0001_initial"
down_revision: Union[str, None] = None
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # IF NOT EXISTS keeps this compatible with installations initially created
    # by database/migrations/001_init.sql.
    op.execute("""
        CREATE TABLE IF NOT EXISTS stories (
            id SERIAL PRIMARY KEY,
            title VARCHAR(255) NOT NULL,
            content TEXT NOT NULL,
            source VARCHAR(100),
            category VARCHAR(50),
            culture VARCHAR(50),
            tags TEXT[],
            created_at TIMESTAMP DEFAULT NOW(),
            processed BOOLEAN DEFAULT FALSE
        )
    """)
    op.execute("""
        CREATE TABLE IF NOT EXISTS videos (
            id SERIAL PRIMARY KEY,
            story_id INTEGER REFERENCES stories(id),
            file_path VARCHAR(500),
            duration INTEGER,
            format VARCHAR(20),
            resolution VARCHAR(20),
            status VARCHAR(50) DEFAULT 'pending',
            metadata JSONB,
            created_at TIMESTAMP DEFAULT NOW()
        )
    """)
    op.execute("CREATE INDEX IF NOT EXISTS idx_videos_status ON videos(status)")
    op.execute("CREATE INDEX IF NOT EXISTS idx_videos_story_id ON videos(story_id)")


def downgrade() -> None:
    op.execute("DROP TABLE IF EXISTS videos")
    op.execute("DROP TABLE IF EXISTS stories")
