package main

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	videoExchange   = "video_events"
	processRouteKey = "video.process"
)

type VideoProcessingMessage struct {
	VideoID   uint  `json:"video_id"`
	StoryID   uint  `json:"story_id"`
	Timestamp int64 `json:"timestamp"`
}

type RabbitMQPublisher struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

func NewRabbitMQPublisher(url string) (*RabbitMQPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := channel.ExchangeDeclare(videoExchange, "topic", true, false, false, false, nil); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return nil, err
	}
	return &RabbitMQPublisher{connection: conn, channel: channel}, nil
}

func (p *RabbitMQPublisher) PublishVideoReadyForProcessing(videoID, storyID uint) error {
	body, err := json.Marshal(VideoProcessingMessage{VideoID: videoID, StoryID: storyID, Timestamp: time.Now().Unix()})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.channel.PublishWithContext(ctx, videoExchange, processRouteKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

func (p *RabbitMQPublisher) Close() {
	if p == nil {
		return
	}
	if p.channel != nil {
		_ = p.channel.Close()
	}
	if p.connection != nil {
		_ = p.connection.Close()
	}
}
