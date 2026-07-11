package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Video struct {
	ID          uint      `gorm:"primaryKey"`
	StoryID     uint      `gorm:"column:story_id"`
	FilePath    string    `gorm:"column:file_path"`
	Duration    int       `gorm:"column:duration"`
	Format      string    `gorm:"column:format"`
	Resolution  string    `gorm:"column:resolution"`
	Status      string    `gorm:"column:status"` // pending, processing, ready, failed
	Metadata    string    `gorm:"column:metadata"` // JSON string
	CreatedAt   time.Time `gorm:"column:created_at"`
}

type Story struct {
	ID        uint      `gorm:"primaryKey"`
	Title     string    `gorm:"column:title"`
	Content   string    `gorm:"column:content"`
	Source    string    `gorm:"column:source"`
	Category  string    `gorm:"column:category"`
	Culture   string    `gorm:"column:culture"`
	Tags      string    `gorm:"column:tags"` // JSON array
	CreatedAt time.Time `gorm:"column:created_at"`
	Processed bool      `gorm:"column:processed"`
}

type InstagramPost struct {
	ID              uint      `gorm:"primaryKey"`
	VideoID         uint      `gorm:"column:video_id"`
	InstagramID     string    `gorm:"column:instagram_id"`
	Caption         string    `gorm:"column:caption"`
	Hashtags        string    `gorm:"column:hashtags"` // JSON array
	PostedAt        time.Time `gorm:"column:posted_at"`
	EngagementMetrics string    `gorm:"column:engagement_metrics"` // JSON
	Status          string    `gorm:"column:status"` // scheduled, posted, failed
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

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Declare exchange
	err = ch.ExchangeDeclare(
		"video_events", // name
		"topic",        // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &RabbitMQPublisher{
		connection: conn,
		channel:    ch,
	}, nil
}

func (p *RabbitMQPublisher) PublishVideoReadyForProcessing(videoID uint, storyID uint) error {
	message := VideoProcessingMessage{
		VideoID:   videoID,
		StoryID:   storyID,
		Timestamp: time.Now().Unix(),
	}

	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.channel.PublishWithContext(
		ctx,
		"video_events",    // exchange
		"video.process",   // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

func (p *RabbitMQPublisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.connection != nil {
		p.connection.Close()
	}
}

type VideoProcessingMessage struct {
	VideoID   uint64 `json:"video_id"`
	StoryID   uint64 `json:"story_id"`
	Timestamp int64  `json:"timestamp"`
}

type Server struct {
	Router      *gin.Engine
	DB          *gorm.DB
	RabbitMQ    *RabbitMQPublisher
	rabbitURL   string
}

func NewServer() *Server {
	// Load environment variables
	godotenv.Load()

	// Initialize Gin
	router := gin.Default()

	// Database connection
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	dsn := "host=" + dbHost + " user=" + dbUser + " password=" + dbPassword +
		" dbname=" + dbName + " port=" + dbPort + " sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	// Auto migrate schemas
	db.AutoMigrate(&Video{}, &Story{}, &InstagramPost{})

	// RabbitMQ setup
	rabbitUser := os.Getenv("RABBITMQ_USER")
	rabbitPass := os.Getenv("RABBITMQ_PASS")
	rabbitHost := os.Getenv("RABBITMQ_HOST")
	rabbitPort := os.Getenv("RABBITMQ_PORT")
	rabbitURL := "amqp://" + rabbitUser + ":" + rabbitPass + "@" + rabbitHost + ":" + rabbitPort + "/"

	rabbitMQ, err := NewRabbitMQPublisher(rabbitURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to RabbitMQ: %v", err)
		// Continue without RabbitMQ for now - graceful degradation
		rabbitMQ = nil
	}

	return &Server{
		Router:   router,
		DB:       db,
		RabbitMQ: rabbitMQ,
		rabbitURL: rabbitURL,
	}
}

func (s *Server) SetupRoutes() {
	// Health check
	s.Router.GET("/health", s.HealthCheck)

	// Video management
	s.Router.GET("/videos", s.GetVideos)
	s.Router.POST("/videos", s.CreateVideo)
	s.Router.GET("/videos/:id", s.GetVideoByID)
	s.Router.PUT("/videos/:id", s.UpdateVideo)
	s.Router.DELETE("/videos/:id", s.DeleteVideo)

	// Video queue management
	s.Router.POST("/queue", s.AddToQueue)
	s.Router.GET("/queue", s.GetQueue)
	s.Router.POST("/process", s.ProcessQueue)

	// Statistics
	s.Router.GET("/stats", s.GetStats)
}

func (s *Server) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now().Unix()})
}

func (s *Server) GetVideos(c *gin.Context) {
	var videos []Video
	if err := s.DB.Find(&videos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, videos)
}

func (s *Server) CreateVideo(c *gin.Context) {
	var video Video
	if err := c.ShouldBindJSON(&video); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.DB.Create(&video).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, video)
}

func (s *Server) GetVideoByID(c *gin.Context) {
	id := c.Param("id")
	var video Video
	if err := s.DB.First(&video, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}
	c.JSON(http.StatusOK, video)
}

func (s *Server) UpdateVideo(c *gin.Context) {
	id := c.Param("id")
	var video Video
	if err := s.DB.First(&video, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	if err := c.ShouldBindJSON(&video); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.DB.Model(&video).Updates(video).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, video)
}

func (s *Server) DeleteVideo(c *gin.Context) {
	id := c.Param("id")
	if err := s.DB.Delete(&Video{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Video deleted"})
}

func (s *Server) AddToQueue(c *gin.Context) {
	var request struct {
		StoryID uint `json:"story_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if story exists and is processed
	var story Story
	if err := s.DB.First(&story, request.StoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Story not found"})
		return
	}
	if !story.Processed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Story not processed yet"})
		return
	}

	// Check if video already exists for this story
	var existing Video
	if err := s.DB.Where("story_id = ?", request.StoryID).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Video already exists for this story"})
		return
	}

	// Create video entry with pending status
	video := Video{
		StoryID: request.StoryID,
		Status:  "pending",
	}

	if err := s.DB.Create(&video).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Publish message to RabbitMQ if available
	if s.RabbitMQ != nil {
		go func() {
			if err := s.RabbitMQ.PublishVideoReadyForProcessing(video.ID, story.ID); err != nil {
				log.Printf("Failed to publish RabbitMQ message: %v", err)
			}
		}()
	}

	c.JSON(http.StatusCreated, video)
}

func (s *Server) GetQueue(c *gin.Context) {
	var videos []Video
	if err := s.DB.Where("status IN (?)", []string{"pending", "processing"}).Find(&videos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, videos)
}

func (s *Server) ProcessQueue(c *gin.Context) {
	// Find a pending video to process
	var video Video
	if err := s.DB.Where("status = ?", "pending").First(&video).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "No videos in queue"})
		return
	}

	// Update status to processing
	video.Status = "processing"
	if err := s.DB.Save(&video).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// In a real implementation, this would trigger the video generation process
	// For now, we'll simulate processing
	go func(videoID uint) {
		// Simulate processing time
		time.Sleep(5 * time.Second)

		// Update video status (in reality, this would come from video generation service)
		var video Video
		if err := s.DB.First(&video, videoID).Error; err == nil {
			video.Status = "ready"
			video.FilePath = "/videos/video_" + strconv.FormatUint(uint64(videoID), 10) + ".mp4"
			video.Duration = 45 // seconds
			video.Format = "mp4"
			video.Resolution = "1080x1920"
			metadata := map[string]interface{}{
				"generated_at": time.Now().Unix(),
				"has_audio":    true,
				"has_subtitles": false,
			}
			metadataJSON, _ := json.Marshal(metadata)
			video.Metadata = string(metadataJSON)
			s.DB.Save(&video)
		}
	}(video.ID)

	c.JSON(http.StatusOK, gin.H{"message": "Processing started", "video_id": video.ID})
}

func (s *Server) GetStats(c *gin.Context) {
	var stats struct {
		TotalStories    int64 `json:"total_stories"`
		ProcessedStories int64 `json:"processed_stories"`
		TotalVideos     int64 `json:"total_videos"`
		PendingVideos   int64 `json:"pending_videos"`
		ReadyVideos     int64 `json:"ready_videos"`
		PostedVideos    int64 `json:"posted_videos"`
	}

	s.DB.Model(&Story{}).Count(&stats.TotalStories)
	s.DB.Model(&Story{}).Where("processed = ?", true).Count(&stats.ProcessedStories)
	s.DB.Model(&Video{}).Count(&stats.TotalVideos)
	s.DB.Model(&Video{}).Where("status = ?", "pending").Count(&stats.PendingVideos)
	s.DB.Model(&Video{}).Where("status = ?", "ready").Count(&stats.ReadyVideos)

	// Count posted videos (via Instagram posts)
	s.DB.Model(&InstagramPost{}).Where("status = ?", "posted").Count(&stats.PostedVideos)

	c.JSON(http.StatusOK, stats)
}

func (s *Server) Start() {
	s.SetupRoutes()
	port := os.Getenv("PORT")
	if port == "" {
		port = "9001"
	}
	log.Printf("Server starting on port %s", port)
	if err := s.Router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}

func main() {
	server := NewServer()
	server.Start()
}