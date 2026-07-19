import asyncio
import json
import logging
import os
import threading
from typing import Optional

import pika

logger = logging.getLogger(__name__)
EXCHANGE = "video_events"
PROCESS_KEY = "video.process"
READY_KEY = "video.ready"
QUEUE = "ai_pipeline_video_generation"


def rabbit_url() -> str:
    explicit = os.getenv("RABBITMQ_URL")
    if explicit:
        return explicit
    user = os.getenv("RABBITMQ_USER", "guest")
    password = os.getenv("RABBITMQ_PASS", "guest")
    host = os.getenv("RABBITMQ_HOST", "rabbitmq")
    port = os.getenv("RABBITMQ_PORT", "5672")
    return f"amqp://{user}:{password}@{host}:{port}/"


def publish_video_ready(video_id: int, story_id: int) -> None:
    connection = pika.BlockingConnection(pika.URLParameters(rabbit_url()))
    try:
        channel = connection.channel()
        channel.exchange_declare(exchange=EXCHANGE, exchange_type="topic", durable=True)
        body = json.dumps({"video_id": video_id, "story_id": story_id}).encode()
        channel.basic_publish(
            exchange=EXCHANGE,
            routing_key=READY_KEY,
            body=body,
            properties=pika.BasicProperties(content_type="application/json", delivery_mode=2),
        )
    finally:
        connection.close()


class VideoJobConsumer:
    def __init__(self) -> None:
        self._stop = threading.Event()
        self._thread: Optional[threading.Thread] = None
        self.connected = False

    def start(self) -> None:
        if self._thread and self._thread.is_alive():
            return
        self._thread = threading.Thread(target=self._run, name="video-job-consumer", daemon=True)
        self._thread.start()

    def stop(self) -> None:
        self._stop.set()
        if self._thread:
            self._thread.join(timeout=5)

    def _run(self) -> None:
        while not self._stop.is_set():
            connection = None
            try:
                connection = pika.BlockingConnection(pika.URLParameters(rabbit_url()))
                channel = connection.channel()
                channel.exchange_declare(exchange=EXCHANGE, exchange_type="topic", durable=True)
                channel.queue_declare(queue=QUEUE, durable=True)
                channel.queue_bind(queue=QUEUE, exchange=EXCHANGE, routing_key=PROCESS_KEY)
                channel.basic_qos(prefetch_count=1)
                self.connected = True

                for method, _, body in channel.consume(QUEUE, inactivity_timeout=1, auto_ack=False):
                    if self._stop.is_set():
                        break
                    if method is None:
                        continue
                    try:
                        payload = json.loads(body)
                        story_id = int(payload["story_id"])
                        video_id = int(payload["video_id"])
                        from app.videos import generate_video_for_story

                        success = asyncio.run(generate_video_for_story(story_id, video_id))
                        if success:
                            channel.basic_ack(method.delivery_tag)
                        else:
                            channel.basic_nack(method.delivery_tag, requeue=False)
                    except Exception:
                        logger.exception("Video job failed")
                        channel.basic_nack(method.delivery_tag, requeue=False)
            except Exception as exc:
                self.connected = False
                if not self._stop.is_set():
                    logger.warning("RabbitMQ consumer unavailable; retrying: %s", exc)
                    self._stop.wait(5)
            finally:
                self.connected = False
                if connection and connection.is_open:
                    try:
                        connection.close()
                    except Exception:
                        pass


video_consumer = VideoJobConsumer()
