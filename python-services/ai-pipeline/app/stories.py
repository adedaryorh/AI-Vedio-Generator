from fastapi import APIRouter, Query, BackgroundTasks, HTTPException
from typing import List, Optional, Dict, Any
import httpx
import asyncio
import os
import logging
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
from sqlalchemy import text
import openai
from dotenv import load_dotenv

load_dotenv(".env.local", override=True)
load_dotenv()

router = APIRouter(tags=["stories"])

# Database setup
DATABASE_URL = os.getenv("DATABASE_URL", "postgresql+asyncpg://storybot:storybotpass@postgres:5432/storybotdb")
engine = create_async_engine(DATABASE_URL)
AsyncSessionLocal = sessionmaker(engine, class_=AsyncSession, expire_on_commit=False)

# OpenAI setup
openai.api_key = os.getenv("OPENAI_API_KEY")

# External API endpoints
GUTENBERG_API_URL = os.getenv("GUTENBERG_API_URL", "https://www.gutenberg.org/")
ISLAMIC_STORIES_API_URL = os.getenv("ISLAMIC_STORIES_API_URL", "https://api.islamicstories.com/")
AFRICAN_FOLKLORE_API_URL = os.getenv("AFRICAN_FOLKLORE_API_URL", "https://api.africanfolklore.com/")

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@router.get("/stories")
async def list_stories(category: Optional[str] = None):
    async with AsyncSessionLocal() as session:
        query = "SELECT id, title, content, source, category, culture, tags, created_at FROM stories WHERE processed = true"
        params = {}
        if category:
            query += " AND category = :category"
            params["category"] = category

        result = await session.execute(text(query), params)
        stories = []
        for row in result:
            stories.append({
                "id": row[0],
                "title": row[1],
                "content": row[2],
                "source": row[3],
                "category": row[4],
                "culture": row[5],
                "tags": row[6] if row[6] is not None else [],
                "created_at": row[7].isoformat() if row[7] is not None else None
            })
        return stories

@router.post("/collect")
async def collect_stories(source: str = Query(..., description="Source to collect from"), background_tasks: BackgroundTasks = None):
    source = source.lower()

    if source == "gutenberg":
        background_tasks.add_task(collect_from_gutenberg)
    elif source == "islamic":
        background_tasks.add_task(collect_from_islamic)
    elif source == "african":
        background_tasks.add_task(collect_from_african)
    elif source == "all":
        background_tasks.add_task(collect_from_gutenberg)
        background_tasks.add_task(collect_from_islamic)
        background_tasks.add_task(collect_from_african)
    else:
        raise HTTPException(status_code=400, detail=f"Unknown source: {source}")

    return {"message": f"Started collecting stories from {source}"}

@router.post("/enhance/{story_id}")
async def enhance_story(story_id: int, background_tasks: BackgroundTasks):
    background_tasks.add_task(enhance_story_with_ai, story_id)
    return {"message": f"Started enhancing story {story_id}"}

async def collect_from_gutenberg():
    try:
        # Fetch some popular books from Gutenberg
        async with httpx.AsyncClient() as client:
            # Get list of popular eBooks
            response = await client.get(f"{GUTENBERG_API_URL}ebooks/search/?sort_order=downloads")
            # For MVP, we'll use some known public domain works
            sample_stories = [
                {
                    "title": "The Tortoise and the Hare",
                    "content": "Once upon a time, in a forest far away, there lived a speedy hare who bragged about how fast he could run. Tired of hearing him boast, Slow and Steady, the tortoise, challenged him to a race. All the animals in the forest gathered to watch. Hare ran down the road for a while and then paused to rest. He looked back at Slow and Steady and cried, 'How do you expect to win this race when you are walking along at your slow, slow pace?' Hare stretched himself out alongside the road and fell asleep, thinking, 'There is plenty of time to relax.' Slow and Steady walked and walked. He never, ever stopped until he came to the finish line. The animals who were watching cheered so loudly for Tortoise, they woke up Hare. Hare stretched and yawned and began to run again, but it was too late. Tortoise was over the line first. After that, Hare always reminded himself, 'Don't brag about your lightning pace, for Slow and Steady won the race!'",
                    "source": "Project Gutenberg",
                    "category": "folklore",
                    "culture": "Western",
                    "tags": ["tortoise", "hare", "fable", "motivational"]
                },
                {
                    "title": "Anansi and the Pot of Wisdom",
                    "content": "Long ago, when the world was young, Anansi the spider decided to collect all the wisdom in the world and keep it for himself. He put all the wisdom into a large pot and said, 'Now I am the wisest of all!' He decided to hide the pot on top of a tall thorny tree where no one could reach it. But try as he might, Anansi couldn't climb the tree with the pot tied in front of him. His young son, seeing his struggle, said, 'Father, why not tie the pot behind you and then try to climb?' Anansi followed his son's advice and climbed the tree easily. When he reached the top, he looked at the pot of wisdom and said, 'I thought I had all the wisdom, yet my own child has just given me a bit more!' In his surprise, he slipped and dropped the pot. The wisdom spilled out and scattered all over the world, so that everyone now has a little bit.",
                    "source": "Project Gutenberg",
                    "category": "folklore",
                    "culture": "African",
                    "tags": ["anansi", "spider", "wisdom", "trickster"]
                }
            ]

            async with AsyncSessionLocal() as session:
                for story in sample_stories:
                    # Check if story already exists
                    result = await session.execute(
                        text("SELECT id FROM stories WHERE title = :title AND source = :source"),
                        {"title": story["title"], "source": story["source"]}
                    )
                    if not result.fetchone():
                        await session.execute(
                            text("""
                                INSERT INTO stories (title, content, source, category, culture, tags, processed)
                                VALUES (:title, :content, :source, :category, :culture, :tags, false)
                            """),
                            {
                                "title": story["title"],
                                "content": story["content"],
                                "source": story["source"],
                                "category": story["category"],
                                "culture": story["culture"],
                                "tags": story["tags"]
                            }
                        )
                await session.commit()

        logger.info(f"Collected {len(sample_stories)} stories from Gutenberg")
    except Exception as e:
        logger.error(f"Error collecting from Gutenberg: {e}")

async def collect_from_islamic():
    try:
        # Sample Islamic stories for MVP
        sample_stories = [
            {
                "title": "The Prophet and the Ants",
                "content": "It is narrated that Prophet Sulaiman (Solomon) was once passing through a valley with his army. As they approached, he heard an ant warning others: 'O ants, enter your dwellings lest Solomon and his armies crush you while they perceive not.' Prophet Sulaiman smiled at the ant's wisdom and ordered his army to take a different path. He thanked the ant for its wisdom and continued on his way, reminding his followers to always be mindful of even the smallest creatures.",
                "source": "Islamic Stories",
                "category": "religious",
                "culture": "Islamic",
                "tags": ["solomon", "ants", "wisdom", "prophet"]
            },
            {
                "title": "The Kindness of Prophet Muhammad",
                "content": "A Bedouin came to Prophet Muhammad and said, 'O Messenger of Allah, I have come to you from a far land.' The Prophet welcomed him and offered him food and drinking water. After the man ate and drank, he said, 'I have come to you to ask about something.' The Prophet said, 'Ask.' The man said, 'Tell me about faith.' The Prophet said, 'Faith is to believe in Allah, His angels, His books, His messengers, the Last Day, and to believe in divine destiny, both the good and the bad thereof.' The man said, 'You have spoken the truth.' Then he said, 'Tell me about Islam.' The Prophet said, 'Islam is to worship Allah and not associate anything with Him, to establish the prescribed prayers, to give the obligatory charity, to fast the month of Ramadan, and to perform the pilgrimage to the House if you have the means.' The man said, 'You have spoken the truth.' Then he said, 'Tell me about Ihsan (excellence).' The Prophet said, 'It is to worship Allah as if you see Him, for if you do not see Him, He sees you.' The man said, 'You have spoken the truth.' Then he said, 'Tell me about the Hour.' The Prophet said, 'The one asked about it knows no more than the questioner.' The man stood up and left. The Prophet's companions asked him, 'Do you know who that man was?' He said, 'That was Jibril (angel Gabriel). He came to teach you your religion.'",
                "source": "Islamic Stories",
                "category": "religious",
                "culture": "Islamic",
                "tags": ["muhammad", "faith", "islam", "wisdom"]
            }
        ]

        async with AsyncSessionLocal() as session:
            for story in sample_stories:
                # Check if exists
                result = await session.execute(
                    text("SELECT id FROM stories WHERE title = :title AND source = :source"),
                    {"title": story["title"], "source": story["source"]}
                )
                if not result.fetchone():
                    await session.execute(
                        text("""
                            INSERT INTO stories (title, content, source, category, culture, tags, processed)
                            VALUES (:title, :content, :source, :category, :culture, :tags, false)
                        """),
                        {
                            "title": story["title"],
                            "content": story["content"],
                            "source": story["source"],
                            "category": story["category"],
                            "culture": story["culture"],
                            "tags": story["tags"]
                        }
                    )
            await session.commit()

        logger.info(f"Collected {len(sample_stories)} stories from Islamic sources")
    except Exception as e:
        logger.error(f"Error collecting from Islamic sources: {e}")

async def collect_from_african():
    try:
        # Sample African folklore for MVP
        sample_stories = [
            {
                "title": "Why the Sun and Moon Live in the Sky",
                "content": "Many years ago, the sun and water were great friends, and both lived on the earth together. The sun used to visit the water often, but the water never returned his visits. One day the sun finally asked the water why he never came to visit him. The water replied, 'Your house is not big enough to accommodate me and all of my people.' The sun promised to build a very big house. He invited the water to visit him the next day. When the water arrived, he called out to the sun to come out and greet him. But the sun shouted, 'I cannot come out now because my house is not ready!' The water waited and waited, but the sun did not come out. After a long wait, the water began to flow away. Just then, the roof of the sun's house began to leak. The water saw this and laughed, saying, 'I told you your house was not big enough!' The water kept flowing and flowed right up into the sky, followed by the sun who was trying to stop the leak in his roof. And that is why the sun and the moon live in the sky today.",
                "source": "African Folklore",
                "category": "folklore",
                "culture": "African",
                "tags": ["sun", "moon", "water", "folklore", "origin"]
            },
            {
                "title": "The Clever Rabbit and the Lion",
                "content": "One day, a lion was very hungry and started hunting for food. He caught a rabbit and was about to eat it when the rabbit said, 'Please, mighty lion, do not eat me. I can show you where there is much more food.' The lion, being curious, asked the rabbit to lead the way. The rabbit took the lion to a deep well and said, 'Look inside the well. There is a huge pile of food there.' The lion looked into the well and saw his own reflection. Thinking it was another lion with food, he jumped into the well to fight him and drowned. The rabbit hopped away safely, having saved himself and many other animals from the lion's hunger.",
                "source": "African Folklore",
                "category": "folklore",
                "culture": "African",
                "tags": ["rabbit", "lion", "cleverness", "trickster", "wisdom"]
            }
        ]

        async with AsyncSessionLocal() as session:
            for story in sample_stories:
                # Check if exists
                result = await session.execute(
                    text("SELECT id FROM stories WHERE title = :title AND source = :source"),
                    {"title": story["title"], "source": story["source"]}
                )
                if not result.fetchone():
                    await session.execute(
                        text("""
                            INSERT INTO stories (title, content, source, category, culture, tags, processed)
                            VALUES (:title, :content, :source, :category, :culture, :tags, false)
                        """),
                        {
                            "title": story["title"],
                            "content": story["content"],
                            "source": story["source"],
                            "category": story["category"],
                            "culture": story["culture"],
                            "tags": story["tags"]
                        }
                    )
            await session.commit()

        logger.info(f"Collected {len(sample_stories)} stories from African sources")
    except Exception as e:
        logger.error(f"Error collecting from African sources: {e}")

async def enhance_story_with_ai(story_id: int):
    try:
        async with AsyncSessionLocal() as session:
            # Get the story
            result = await session.execute(
                text("SELECT id, title, content FROM stories WHERE id = :id AND processed = false"),
                {"id": story_id}
            )
            story = result.fetchone()
            if not story:
                logger.warning(f"Story {story_id} not found or already processed")
                return

            story_id, title, content = story

            # Enhance the story using OpenAI
            prompt = f"""
            Please enhance this story to make it more engaging and suitable for a short video narrative:

            Title: {title}
            Content: {content}

            Please:
            1. Keep the core message and moral intact
            2. Make it more vivid and engaging for video narration
            3. Add descriptive details where appropriate
            4. Keep it appropriate for all ages
            5. Make it approximately 2-3 minutes when read aloud
            6. Return only the enhanced story content, no additional commentary
            """

            response = openai.ChatCompletion.create(
                model="gpt-3.5-turbo",
                messages=[
                    {"role": "system", "content": "You are a skilled storyteller who enhances traditional tales for modern audiences."},
                    {"role": "user", "content": prompt}
                ],
                max_tokens=500,
                temperature=0.7
            )

            enhanced_content = response.choices[0].message.content.strip()

            # Update the story with enhanced content
            await session.execute(
                text("""
                    UPDATE stories
                    SET content = :content, processed = true
                    WHERE id = :id
                """),
                {"content": enhanced_content, "id": story_id}
            )
            await session.commit()

        logger.info(f"Enhanced story {story_id}: {title}")
    except Exception as e:
        logger.error(f"Error enhancing story {story_id}: {e}")
        # Mark as processed even if enhancement failed to avoid infinite retries
        try:
            async with AsyncSessionLocal() as session:
                await session.execute(
                    text("UPDATE stories SET processed = true WHERE id = :id"),
                    {"id": story_id}
                )
                await session.commit()
        except:
            pass
