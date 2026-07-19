package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (s *Server) HealthCheck(c *gin.Context) {
	status := gin.H{"status": "ok", "timestamp": time.Now().Unix(), "database": "connected", "rabbitmq": "disconnected", "redis": "disconnected"}
	if s.RabbitMQ != nil {
		status["rabbitmq"] = "connected"
	}
	if s.Redis != nil {
		status["redis"] = "connected"
	}
	status["platforms"] = gin.H{"instagram": s.config.Instagram.AccessToken != "", "youtube": s.config.YouTube.APIKey != "", "tiktok": s.config.TikTok.AccessToken != ""}
	breakers := gin.H{}
	for _, platform := range []string{"instagram", "youtube", "tiktok"} {
		if breaker := s.CircuitBreakers[platform]; breaker != nil {
			breakers[platform] = breaker.State()
		} else {
			breakers[platform] = "unavailable"
		}
	}
	status["circuit_breakers"] = breakers
	c.JSON(http.StatusOK, status)
}

func (s *Server) GetVideos(c *gin.Context) {
	var videos []Video
	if err := s.DB.Find(&videos).Error; err != nil {
		respondError(c, err)
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
	if video.Status == "" {
		video.Status = StatusPending
	}
	if err := s.DB.Create(&video).Error; err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, video)
}

func (s *Server) GetVideoByID(c *gin.Context) {
	var video Video
	if err := s.DB.First(&video, c.Param("id")).Error; err != nil {
		respondDatabaseLookup(c, err, "Video not found")
		return
	}
	c.JSON(http.StatusOK, video)
}

func (s *Server) UpdateVideo(c *gin.Context) {
	var video Video
	if err := s.DB.First(&video, c.Param("id")).Error; err != nil {
		respondDatabaseLookup(c, err, "Video not found")
		return
	}
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	delete(updates, "id")
	delete(updates, "story_id")
	if err := s.DB.Model(&video).Updates(updates).Error; err != nil {
		respondError(c, err)
		return
	}
	if err := s.DB.First(&video, video.ID).Error; err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, video)
}

func (s *Server) DeleteVideo(c *gin.Context) {
	result := s.DB.Delete(&Video{}, c.Param("id"))
	if result.Error != nil {
		respondError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Video deleted"})
}

func (s *Server) AddToQueue(c *gin.Context) {
	var request struct {
		StoryID uint `json:"story_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var story Story
	if err := s.DB.First(&story, request.StoryID).Error; err != nil {
		respondDatabaseLookup(c, err, "Story not found")
		return
	}
	if !story.Processed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Story not processed yet"})
		return
	}
	var count int64
	if err := s.DB.Model(&Video{}).Where("story_id = ?", request.StoryID).Count(&count).Error; err != nil {
		respondError(c, err)
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Video already exists for this story"})
		return
	}
	video := Video{StoryID: request.StoryID, Status: StatusPending}
	if err := s.DB.Create(&video).Error; err != nil {
		respondError(c, err)
		return
	}
	if s.RabbitMQ != nil {
		go func() {
			if err := s.RabbitMQ.PublishVideoReadyForProcessing(video.ID, story.ID); err != nil {
				log.Printf("publish processing event: %v", err)
			}
		}()
	}
	c.JSON(http.StatusCreated, video)
}

func (s *Server) GetQueue(c *gin.Context) {
	var videos []Video
	if err := s.DB.Where("status IN ?", []string{StatusPending, StatusProcessing}).Find(&videos).Error; err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, videos)
}

func (s *Server) ProcessQueue(c *gin.Context) {
	var video Video
	if err := s.DB.Where("status = ?", StatusPending).First(&video).Error; err != nil {
		respondDatabaseLookup(c, err, "No videos in queue")
		return
	}
	if err := s.DB.Model(&video).Update("status", StatusProcessing).Error; err != nil {
		respondError(c, err)
		return
	}
	go s.completeSimulatedVideo(video.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Processing started", "video_id": video.ID})
}

func (s *Server) completeSimulatedVideo(videoID uint) {
	time.Sleep(5 * time.Second)
	metadata, err := json.Marshal(map[string]interface{}{"generated_at": time.Now().Unix(), "has_audio": true, "has_subtitles": false})
	if err != nil {
		log.Printf("encode video metadata: %v", err)
		return
	}
	updates := map[string]interface{}{"status": StatusReady, "file_path": fmt.Sprintf("/videos/video_%d.mp4", videoID), "duration": 45, "format": "mp4", "resolution": "1080x1920", "metadata": string(metadata)}
	if err := s.DB.Model(&Video{}).Where("id = ?", videoID).Updates(updates).Error; err != nil {
		log.Printf("complete simulated video %d: %v", videoID, err)
	}
}

func (s *Server) GetStats(c *gin.Context) {
	var stats Stats
	queries := []struct {
		model  interface{}
		query  string
		value  interface{}
		target *int64
	}{
		{&Story{}, "", nil, &stats.TotalStories}, {&Story{}, "processed = ?", true, &stats.ProcessedStories},
		{&Video{}, "", nil, &stats.TotalVideos}, {&Video{}, "status = ?", StatusPending, &stats.PendingVideos},
		{&Video{}, "status = ?", StatusReady, &stats.ReadyVideos}, {&InstagramPost{}, "status = ?", "posted", &stats.PostedVideos},
	}
	for _, item := range queries {
		query := s.DB.Model(item.model)
		if item.query != "" {
			query = query.Where(item.query, item.value)
		}
		if err := query.Count(item.target).Error; err != nil {
			respondError(c, err)
			return
		}
	}
	c.JSON(http.StatusOK, stats)
}

func respondDatabaseLookup(c *gin.Context, err error, message string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": message})
		return
	}
	respondError(c, err)
}

func respondError(c *gin.Context, err error) {
	log.Printf("request failed: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
}
