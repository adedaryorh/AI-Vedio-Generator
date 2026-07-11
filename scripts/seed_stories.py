import os
import sqlalchemy
from sqlalchemy import create_engine, text

DATABASE_URL = os.getenv('DATABASE_URL', 'postgresql+psycopg2://storybot:storybotpass@localhost:5432/storybotdb')
engine = create_engine(DATABASE_URL)

EXAMPLE_STORIES = [
    {
        'title': 'The Clever Tortoise',
        'content': 'Once upon a time in Africa, a clever tortoise outwitted a boastful elephant...',
        'source': 'African Folklore',
        'category': 'folklore',
        'culture': 'African',
        'tags': ['tortoise', 'folklore', 'africa'],
    },
    {
        'title': 'The Prophet and the Ants',
        'content': 'A story from the Quran about the Prophet Solomon and the ants...',
        'source': 'Islamic Stories',
        'category': 'religious',
        'culture': 'Islamic',
        'tags': ['quran', 'prophet', 'islam'],
    },
]

def seed():
    with engine.connect() as conn:
        for story in EXAMPLE_STORIES:
            conn.execute(text('''
                INSERT INTO stories (title, content, source, category, culture, tags, processed)
                VALUES (:title, :content, :source, :category, :culture, :tags, false)
            '''), story)
        conn.commit()

if __name__ == '__main__':
    seed() 