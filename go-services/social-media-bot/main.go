package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

// SocialMediaPost represents a post on any social media platform
type SocialMediaPost struct {
	ID              int       `json:"id"`
	VideoID         int       `json:"video_id"`
	Platform        string    `json:"platform"` // instagram, youtube, tiktok
	PlatformID      string    `json:"platform_id"` // ID from the platform (e.g., Instagram ID, YouTube video ID)
	Caption         string    `json:"caption"`
	Hashtags        []string  `json:"hashtags"`
	PostedAt        time.Time `json:"posted_at"`
	EngagementMetrics map[string]interface{} `json:"engagement_metrics"`
	Status          string    `json:"status"` // scheduled, posting, posted, failed
	PlatformSpecific  interface{} `json:"platform_specific,omitempty"` // Platform-specific data
}

// VideoInfo represents basic video information
type VideoInfo struct {
	ID      int    `json:"id"`
	FilePath string `json:"file_path"`
	Title   string `json:"title"`
}

// VideoProcessingMessage represents a message from the video processing service
type VideoProcessingMessage struct {
	VideoID   uint64 `json:"video_id"`
	StoryID   uint64 `json:"story_id"`
	Timestamp int64  `json:"timestamp"`
}

// SocialMediaServer handles all social media operations
type SocialMediaServer struct {
	router      *http.ServeMux
	rabbitMQ    *RabbitMQConsumer
	wg          sync.WaitGroup
}

// RabbitMQConsumer handles consuming messages from RabbitMQ
type RabbitMQConsumer struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	queueName  string
}

// NewRabbitMQConsumer creates a new RabbitMQ consumer
func NewRabbitMQConsumer(url string, queueName string) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare queue
	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &RabbitMQConsumer{
		connection: conn,
		channel:    ch,
		queueName:  queueName,
	}, nil
}

// Consume starts consuming messages from the queue
func (c *RabbitMQConsumer) Consume(handler func(VideoProcessingMessage) error) error {
	msgs, err := c.channel.Consume(
		c.queueName, // queue
		"",          // consumer
		false,       // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			var message VideoProcessingMessage
			if err := json.Unmarshal(d.Body, &message); err != nil {
				log.Printf("Error decoding message: %v", err)
				d.Nack(false, false) // Don't requeue malformed messages
				continue
			}

			log.Printf("Received video processing message: VideoID=%d, StoryID=%d", message.VideoID, message.StoryID)

			// Process the video (in a real implementation, this would trigger social media posting)
			if err := handler(message); err != nil {
				log.Printf("Error processing video message: %v", err)
				d.Nack(false, true) // Requeue the message
				continue
			}

			d.Ack(false) // Acknowledge the message
		}
	}()

	return nil
}

// Close closes the RabbitMQ connection and channel
func (c *RabbitMQConsumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.connection != nil {
		c.connection.Close()
	}
}

func NewSocialMediaServer() *SocialMediaServer {
	// Load environment variables
	godotenv.Load()

	mux := http.NewServeMux()
	server := &SocialMediaServer{router: mux}

	// Initialize RabbitMQ consumer
	rabbitUser := os.Getenv("RABBITMQ_USER")
	rabbitPass := os.Getenv("RABBITMQ_PASS")
	rabbitHost := os.Getenv("RABBITMQ_HOST")
	rabbitPort := os.Getenv("RABBITMQ_PORT")
	rabbitURL := "amqp://" + rabbitUser + ":" + rabbitPass + "@" + rabbitHost + ":" + rabbitPort + "/"

	rabbitMQ, err := NewRabbitMQConsumer(rabbitURL, "video.ready_for_posting")
	if err != nil {
		log.Printf("Warning: Failed to connect to RabbitMQ: %v. Social media posting will only work via HTTP endpoints.", err)
		// Continue without RabbitMQ for graceful degradation
		server.rabbitMQ = nil
	} else {
		server.rabbitMQ = rabbitMQ
		// Start consuming messages
		server.wg.Add(1)
		go func() {
			defer server.wg.Done()
			if err := server.rabbitMQ.Consume(server.processVideoMessage); err != nil {
				log.Printf("Error starting RabbitMQ consumer: %v", err)
			}
		}()
	}

	return server
}

// processVideoMessage handles incoming video processing messages from RabbitMQ
func (s *SocialMediaServer) processVideoMessage(message VideoProcessingMessage) error {
	log.Printf("Processing video %d for social media posting", message.VideoID)

	// Get video info (in reality, this would call content-manager service)
	videoInfo := s.getVideoInfo(int(message.VideoID))
	if videoInfo.ID == 0 {
		return fmt.Errorf("video not found: %d", message.VideoID)
	}

	// Generate a caption based on the video title
	caption := fmt.Sprintf("Check out this amazing story: %s", videoInfo.Title)

	// Generate hashtags based on video content
	hashtags := s.generateHashtags(videoInfo.Title)

	// Post to all platforms (in a real implementation, you might want to configure this)
	platforms := []string{"instagram", "youtube", "tiktok"}
	for _, platform := range platforms {
		var platformID string
		var err error

		switch platform {
		case "instagram":
			platformID, err = s.simulateInstagramPost(videoInfo.FilePath, caption, hashtags)
		case "youtube":
			platformID, err = s.simulateYouTubeUpload(videoInfo.FilePath, caption, hashtags)
		case "tiktok":
			platformID, err = s.simulateTikTokUpload(videoInfo.FilePath, caption, hashtags)
		}

		if err != nil {
			log.Printf("Failed to post to %s: %v", platform, err)
			continue
		}

		// Record the post (in reality, this would save to database)
		log.Printf("Successfully posted to %s with ID: %s", platform, platformID)
	}

	return nil
}

func (s *SocialMediaServer) SetupRoutes() {
	// Health check
	s.router.HandleFunc("/health", s.HealthCheck)

	// Social media posting (supports multiple platforms)
	s.router.HandleFunc("/post", s.PostToSocialMedia)
	s.router.HandleFunc("/schedule", s.SchedulePost)

	// Analytics
	s.router.HandleFunc("/analytics", s.GetAnalytics)
	s.router.HandleFunc("/posts", s.GetPosts)

	// Platform-specific endpoints
	s.router.HandleFunc("/platforms", s.GetSupportedPlatforms)
}

func (s *SocialMediaServer) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := map[string]string{"status": "ok", "timestamp": fmt.Sprint(time.Now().Unix())}
	if s.rabbitMQ != nil {
		status["rabbitmq"] = "connected"
	} else {
		status["rabbitmq"] = "disconnected"
	}
	json.NewEncoder(w).Encode(status)
}

func (s *SocialMediaServer) PostToSocialMedia(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		VideoID  int    `json:"video_id"`
		Caption  string `json:"caption"`
		Platform string `json:"platform"` // instagram, youtube, tiktok
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Set default platform if not specified
	if request.Platform == "" {
		request.Platform = "instagram"
	}

	// Validate platform
	if !s.isSupportedPlatform(request.Platform) {
		http.Error(w, fmt.Sprintf("Unsupported platform: %s. Supported platforms: instagram, youtube, tiktok", request.Platform), http.StatusBadRequest)
		return
	}

	// Get video info (in reality, this would call content-manager service)
	videoInfo := s.getVideoInfo(request.VideoID)
	if videoInfo.ID == 0 {
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}

	// Generate hashtags based on video content
	hashtags := s.generateHashtags(videoInfo.Title)

	// Post to the selected platform
	var platformID string
	var err error

	switch request.Platform {
	case "instagram":
		platformID, err = s.simulateInstagramPost(videoInfo.FilePath, request.Caption, hashtags)
	case "youtube":
		platformID, err = s.simulateYouTubeUpload(videoInfo.FilePath, request.Caption, hashtags)
	case "tiktok":
		platformID, err = s.simulateTikTokUpload(videoInfo.FilePath, request.Caption, hashtags)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to post to %s: %v", request.Platform, err), http.StatusInternalServerError)
		return
	}

	// Record the post (in reality, this would save to database)
	post := SocialMediaPost{
		VideoID:         request.VideoID,
		Platform:        request.Platform,
		PlatformID:      platformID,
		Caption:         request.Caption,
		Hashtags:        hashtags,
		PostedAt:        time.Now(),
		EngagementMetrics: map[string]interface{}{
			"likes":     0,
			"comments":  0,
			"shares":    0,
			"saves":     0,
			"views":     0, // Added for YouTube/TikTok
		},
		Status: "posted",
	}

	// Add platform-specific data
	switch request.Platform {
	case "youtube":
		post.PlatformSpecific = map[string]string{
			"video_url": fmt.Sprintf("https://youtube.com/watch?v=%s", platformID),
		}
	case "tiktok":
		post.PlatformSpecific = map[string]string{
			"video_url": fmt.Sprintf("https://tiktok.com/@user/video/%s", platformID),
		}
	}

	// In a real app, we'd save this to database
	// For MVP, we'll just return the result

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    fmt.Sprintf("Successfully posted to %s", request.Platform),
		"platform_id": platformID,
		"post":       post,
	})
}

func (s *SocialMediaServer) SchedulePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		VideoID         int       `json:"video_id"`
		Caption         string    `json:"caption"`
		Platform        string    `json:"platform"`
		ScheduleTime    time.Time `json:"schedule_time"`
		Timezone        string    `json:"timezone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Set default platform if not specified
	if request.Platform == "" {
		request.Platform = "instagram"
	}

	// Validate platform
	if !s.isSupportedPlatform(request.Platform) {
		http.Error(w, fmt.Sprintf("Unsupported platform: %s. Supported platforms: instagram, youtube, tiktok", request.Platform), http.StatusBadRequest)
		return
	}

	// Check if video exists
	videoInfo := s.getVideoInfo(request.VideoID)
	if videoInfo.ID == 0 {
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}

	// In a real implementation, we would save this to a database and use a scheduler
	// For MVP, we'll just acknowledge the request

	if time.Now().After(request.ScheduleTime) {
		// If scheduled time is in the past, post immediately
		// We need to create a new request with the same data but without scheduling
		tempReq := struct {
			VideoID  int    `json:"video_id"`
			Caption  string `json:"caption"`
			Platform string `json:"platform"`
		}{
			VideoID:  request.VideoID,
			Caption:  request.Caption,
			Platform: request.Platform,
		}

		// Create a new request body for immediate posting
		reqBody, _ := json.Marshal(tempReq)
		r.Body = io.NopCloser(bytes.NewReader(reqBody))
		r.ContentLength = int64(len(reqBody))
		s.PostToSocialMedia(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     fmt.Sprintf("Post scheduled successfully for %s", request.Platform),
		"video_id":    request.VideoID,
		"platform":    request.Platform,
		"schedule_time": request.ScheduleTime,
		"status":      "scheduled",
	})
}

// GetSupportedPlatforms returns the list of supported social media platforms
func (s *SocialMediaServer) GetSupportedPlatforms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{
		"platforms": {"instagram", "youtube", "tiktok"},
	})
}

func (s *SocialMediaServer) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	// In a real app, this would fetch actual analytics from APIs or database
	// For MVP, we'll return mock data

	analytics := map[string]interface{}{
		"total_posts":      12,
		"total_views":      15420,
		"total_likes":      892,
		"total_comments":   156,
		"average_engagement": 5.8,
		"platform_breakdown": map[string]int{
			"instagram": 5,
			"youtube":   4,
			"tiktok":    3,
		},
		"top_performing_posts": []map[string]interface{}{
			{
				"post_id":   "ig_12345",
				"platform":  "instagram",
				"views":     3200,
				"likes":     210,
				"engagement": 8.2,
			},
			{
				"post_id":   "yt_67890",
				"platform":  "youtube",
				"views":     5400,
				"likes":     320,
				"engagement": 7.8,
			},
			{
				"post_id":   "tk_54321",
				"platform":  "tiktok",
				"views":     12500,
				"likes":     890,
				"engagement": 9.1,
			},
		},
		"growth_rate":    12.5,
		"audience_demographics": map[string]interface{}{
			"age_18_24": 35,
			"age_25_34": 40,
			"age_35_44": 18,
			"age_45_plus": 7,
		},
		"top_countries": []string{"US", "UK", "CA", "AU"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

func (s *SocialMediaServer) GetPosts(w http.ResponseWriter, r *http.Request) {
	// In a real app, this would fetch from database
	// For MVP, return mock data

	posts := []SocialMediaPost{
		{
			ID:              1,
			VideoID:         101,
			Platform:        "instagram",
			PlatformID:      "ig_12345",
			Caption:         "The Tortoise and the Hare - A timeless tale of perseverance",
			Hashtags:        []string{"#TortoiseAndHare", "#FableFriday", "#WisdomWednesday", "#MotivationMonday", "#ShortStory"},
			PostedAt:        time.Now().Add(-48 * time.Hour),
			EngagementMetrics: map[string]interface{}{
				"likes":    124,
				"comments": 18,
				"shares":   7,
				"saves":    23,
			},
			Status: "posted",
		},
		{
			ID:              2,
			VideoID:         102,
			Platform:        "youtube",
			PlatformID:      "yt_67890",
			Caption:         "Why the Sun and Moon Live in the Sky - African Folktale Explained",
			Hashtags:        []string{"#AfricanFolklore", "#Mythology", "#Educational", "#ShortFilm"},
			PostedAt:        time.Now().Add(-36 * time.Hour),
			EngagementMetrics: map[string]interface{}{
				"likes":    320,
				"comments": 45,
				"shares":   22,
				"views":    5400,
			},
			Status: "posted",
			PlatformSpecific: map[string]string{
				"video_url": "https://youtube.com/watch?v=yt_67890",
			},
		},
		{
			ID:              3,
			VideoID:         103,
			Platform:        "tiktok",
			PlatformID:      "tk_54321",
			Caption:         "Quick Wisdom: The Clever Rabbit and the Lion",
			Hashtags:        []string{"#AfricanFolktale", "#WisdomWednesday", "#QuickStory", "#LearnTok"},
			PostedAt:        time.Now().Add(-12 * time.Hour),
			EngagementMetrics: map[string]interface{}{
				"likes":    890,
				"comments": 120,
				"shares":   230,
				"views":    12500,
			},
			Status: "posted",
			PlatformSpecific: map[string]string{
				"video_url": "https://tiktok.com/@storybot/video/tk_54321",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// Helper functions

func (s *SocialMediaServer) isSupportedPlatform(platform string) bool {
	switch strings.ToLower(platform) {
	case "instagram", "youtube", "tiktok":
		return true
	default:
		return false
	}
}

func (s *SocialMediaServer) getVideoInfo(videoID int) VideoInfo {
	// In a real implementation, this would call the content-manager service
	// For MVP, return mock data

	if videoID == 101 {
		return VideoInfo{
			ID:      101,
			FilePath: "/app/videos/tortoise_and_hare.mp4",
			Title:   "The Tortoise and the Hare",
		}
	} else if videoID == 102 {
		return VideoInfo{
			ID:      102,
			FilePath: "/app/videos/sun_and_moon.mp4",
			Title:   "Why the Sun and Moon Live in the Sky",
		}
	} else if videoID == 103 {
		return VideoInfo{
			ID:      103,
			FilePath: "/app/videos/clever_rabbit_lion.mp4",
			Title:   "The Clever Rabbit and the Lion",
		}
	}
	return VideoInfo{}
}

func (s *SocialMediaServer) generateHashtags(title string) []string {
	// Generate relevant hashtags based on video title
	baseTags := []string{"#ShortStory", "#StoryTime"}

	titleLower := strings.ToLower(title)

	// Common story categories
	if strings.Contains(titleLower, "tortoise") || strings.Contains(titleLower, "hare") {
		return append(baseTags, "#TortoiseAndHare", "#Fable", "#Perseverance", "#Motivation")
	} else if strings.Contains(titleLower, "sun") && strings.Contains(titleLower, "moon") {
		return append(baseTags, "#AfricanFolklore", "#SunAndMoon", "#OriginStory", "#Mythology", "#Culture")
	} else if strings.Contains(titleLower, "prophet") || strings.Contains(titleLower, "solomon") {
		return append(baseTags, "#IslamicStory", "#ProphetStories", "#Wisdom", "#Faith")
	} else if strings.Contains(titleLower, "rabbit") || strings.Contains(titleLower, "lion") {
		return append(baseTags, "#AfricanFolktale", "#CleverRabbit", "#TricksterTales", "#Wisdom")
	} else {
		// Generic tags
		return append(baseTags, "#StoryOfTheDay", "#BedtimeStory", "#Inspiration")
	}
}

func (s *SocialMediaServer) simulateInstagramPost(videoPath, caption string, hashtags []string) (string, error) {
	// In a real implementation, this would use the Instagram Graph API
	// For MVP, we'll simulate the post

	// Simulate network delay
	time.Sleep(2 * time.Second)

	// Generate a fake Instagram ID
	postID := fmt.Sprintf("ig_%d", time.Now().Unix())

	log.Printf("Simulating Instagram post:")
	log.Printf("  Video: %s", videoPath)
	log.Printf("  Caption: %s", caption)
	log.Printf("  Hashtags: %v", hashtags)
	log.Printf("  Instagram ID: %s", postID)

	return postID, nil
}

func (s *SocialMediaServer) simulateYouTubeUpload(videoPath, title string, hashtags []string) (string, error) {
	// In a real implementation, this would use the YouTube Data API
	// For MVP, we'll simulate the upload

	// Simulate network delay and processing time
	time.Sleep(3 * time.Second)

	// Generate a fake YouTube video ID
	videoID := fmt.Sprintf("yt_%d", time.Now().Unix())

	log.Printf("Simulating YouTube upload:")
	log.Printf("  Video: %s", videoPath)
	log.Printf("  Title: %s", title)
	log.Printf("  Hashtags: %v", hashtags)
	log.Printf("  YouTube Video ID: %s", videoID)

	return videoID, nil
}

func (s *SocialMediaServer) simulateTikTokUpload(videoPath, caption string, hashtags []string) (string, error) {
	// In a real implementation, this would use the TikTok API
	// For MVP, we'll simulate the upload

	// Simulate network delay and processing time
	time.Sleep(2 * time.Second)

	// Generate a fake TikTok video ID
	videoID := fmt.Sprintf("tk_%d", time.Now().Unix())

	log.Printf("Simulating TikTok upload:")
	log.Printf("  Video: %s", videoPath)
	log.Printf("  Caption: %s", caption)
	log.Printf("  Hashtags: %v", hashtags)
	log.Printf("  TikTok Video ID: %s", videoID)

	return videoID, nil
}

func (s *SocialMediaServer) Start() {
	s.SetupRoutes()
	port := os.Getenv("PORT")
	if port == "" {
		port = "9002"
	}
	log.Printf("Social Media Bot server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, s.router); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}

// Wait waits for all goroutines to finish
func (s *SocialMediaServer) Wait() {
	s.wg.Wait()
	// Close RabbitMQ connection
	if s.rabbitMQ != nil {
		s.rabbitMQ.Close()
	}
}

func main() {
	server := NewSocialMediaServer()
	go server.Start()

	// Wait for interrupt signal to gracefully shut down
	// For simplicity, we're just blocking here
	// In a real app, you'd use signal handling
	select {}
}