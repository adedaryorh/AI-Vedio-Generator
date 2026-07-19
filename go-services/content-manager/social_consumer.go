package main

import (
	"encoding/json"
	"log"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

type videoReadyMessage struct {
	VideoID uint `json:"video_id"`
	StoryID uint `json:"story_id"`
}

func (s *Server) startSocialConsumer() {
	if s.RabbitMQ == nil || s.RabbitMQ.connection == nil {
		return
	}
	channel, err := s.RabbitMQ.connection.Channel()
	if err != nil {
		log.Printf("social consumer channel: %v", err)
		return
	}
	queue, err := channel.QueueDeclare("content_manager_social_publish", true, false, false, false, nil)
	if err != nil {
		_ = channel.Close()
		log.Printf("social queue: %v", err)
		return
	}
	for _, key := range []string{"video.ready_for_posting", "video.ready"} {
		if err := channel.QueueBind(queue.Name, key, videoExchange, false, nil); err != nil {
			log.Printf("social queue bind %s: %v", key, err)
		}
	}
	deliveries, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		_ = channel.Close()
		log.Printf("social consume: %v", err)
		return
	}
	go func() {
		defer channel.Close()
		for delivery := range deliveries {
			var message videoReadyMessage
			if err := json.Unmarshal(delivery.Body, &message); err != nil {
				_ = delivery.Nack(false, false)
				continue
			}
			if err := s.autoPublishReadyVideo(message.VideoID); err != nil {
				log.Printf("automatic social publish video %d: %v", message.VideoID, err)
				_ = delivery.Nack(false, true)
				continue
			}
			_ = delivery.Ack(false)
		}
	}()
}

func (s *Server) autoPublishReadyVideo(videoID uint) error {
	platforms := splitCSV(valueOrDefault("AUTO_POST_PLATFORMS", "instagram"))
	for _, platform := range platforms {
		platform = strings.ToLower(strings.TrimSpace(platform))
		if platform == "" {
			continue
		}
		if _, err := s.publish(socialPostRequest{VideoID: videoID, Platform: platform, Caption: "New story video"}); err != nil {
			return err
		}
	}
	return nil
}

var _ amqp.Delivery
