package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

// SocialMediaServer represents the social media bot service
type SocialMediaServer struct {
	router          *mux.Router
	rabbitMQ        *amqp091.Connection
	instagramCB     *CircuitBreaker
	youtubeCB       *CircuitBreaker
	tiktokCB        *CircuitBreaker
	redisClient     *redis.Client
	rateLimiter     *rate.Limiter
	rateLimitConfig RateLimitConfig
	instagramConfig InstagramConfig
	youtubeConfig   YouTubeConfig
	tiktokConfig    TikTokConfig
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond float64
	Burst             int
	Enabled           bool
}

// InstagramConfig holds Instagram API configuration
type InstagramConfig struct {
	AppID       string
	AppSecret   string
	AccessToken string
	Enabled     bool
	MaxRetries  int
}

// YouTubeConfig holds YouTube API configuration
type YouTubeConfig struct {
	APIKey    string
	Enabled   bool
	MaxRetries int
}

// TikTokConfig holds TikTok API configuration
type TikTokConfig struct {
	ClientKey    string
	ClientSecret string
	AccessToken  string
	Enabled      bool
	MaxRetries   int
}

// Video represents a video entity
type Video struct {
	ID        int
	Title     string
	FilePath  string
}

// SocialMediaPost represents a social media post
type SocialMediaPost struct {
	VideoID         int
	Platform        string
	PlatformID      string
	Caption         string
	Hashtags        []string
	PostedAt        time.Time
	EngagementMetrics map[string]interface{}
	PlatformSpecific  map[string]string
	Status          string
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	Closed    CircuitBreakerState = "closed"
	Open      CircuitBreakerState = "open"
	HalfOpen  CircuitBreakerState = "half_open"
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu sync.RWMutex
	state CircuitBreakerState
	failureCount int
	maxFailures int
	timeout time.Duration
	lastFailureTime time.Time
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       Closed,
		failureCount: 0,
		maxFailures: maxFailures,
		timeout:     timeout,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	cb.mu.Lock()
	if cb.state == Open {
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.mu.Unlock()
			cb.mu.Lock()
			cb.state = HalfOpen
			cb.mu.Unlock()
		} else {
			cb.mu.Unlock()
			return nil, fmt.Errorf("circuit breaker is open")
		}
	}
	cb.mu.Unlock()

	result, err := fn()
	if err != nil {
		cb.recordFailure()
		return nil, err
	}
	cb.recordSuccess()
	return result, nil
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	if cb.failureCount >= cb.maxFailures {
		cb.state = Open
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount = 0
	cb.state = Closed
}

// NewSocialMediaServer creates a new social media server instance
func NewSocialMediaServer() *SocialMediaServer {
	router := mux.NewRouter()

	// Default rate limit config (can be overridden by environment)
	rateLimitConfig := RateLimitConfig{
		RequestsPerSecond: 5.0, // 5 requests per second
		Burst:             10,  // Allow bursts up to 10 requests
		Enabled:           true,
	}

	// Create rate limiter
	var limiter *rate.Limiter
	if rateLimitConfig.Enabled {
		limiter = rate.NewLimiter(rate.Limit(rateLimitConfig.RequestsPerSecond), rateLimitConfig.Burst)
	} else {
		// Disable rate limiting by setting a very high limit
		limiter = rate.NewLimiter(rate.Inf, 0)
	}

	return &SocialMediaServer{
		router:          router,
		instagramCB:     NewCircuitBreaker(5, 60*time.Second),
		youtubeCB:       NewCircuitBreaker(5, 60*time.Second),
		tiktokCB:        NewCircuitBreaker(5, 60*time.Second),
		rateLimiter:     limiter,
		rateLimitConfig: rateLimitConfig,
	}
}

// SetupRoutes sets up all HTTP endpoints
func (s *SocialMediaServer) SetupRoutes() {
	// Health check
	s.router.HandleFunc("/health", s.HealthCheck)

	// Metrics endpoint
	s.router.HandleFunc("/metrics", s.MetricsHandler)

	// Social media posting (supports multiple platforms) - with rate limiting and metrics
	s.router.HandleFunc("/post", s.metricsMiddleware(s.rateLimitMiddleware(s.PostToSocialMedia)))
	s.router.HandleFunc("/schedule", s.metricsMiddleware(s.rateLimitMiddleware(s.SchedulePost)))

	// Analytics - with rate limiting and metrics
	s.router.HandleFunc("/analytics", s.metricsMiddleware(s.rateLimitMiddleware(s.GetAnalytics)))
	s.router.HandleFunc("/posts", s.metricsMiddleware(s.rateLimitMiddleware(s.GetPosts)))

	// Platform-specific endpoints - with rate limiting and metrics
	s.router.HandleFunc("/platforms", s.metricsMiddleware(s.rateLimitMiddleware(s.GetSupportedPlatforms)))
}

// HealthCheck returns the health status of the service
func (s *SocialMediaServer) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := map[string]string{"status": "ok", "timestamp": fmt.Sprint(time.Now().Unix())}

	if s.rabbitMQ != nil {
		status["rabbitmq"] = "connected"
	} else {
		status["rabbitmq"] = "disconnected"
	}

	// Add platform status
	status["instagram_api"] = boolToString(s.instagramConfig.Enabled)
	status["youtube_api"] = boolToString(s.youtubeConfig.Enabled)
	status["tiktok_api"] = boolToString(s.tiktokConfig.Enabled)

	// Add circuit breaker status
	status["instagram_cb"] = string(s.instagramCB.State())
	status["youtube_cb"] = string(s.youtubeCB.State())
	status["tiktok_cb"] = string(s.tiktokCB.State())

	// Add Redis status
	if s.redisClient != nil {
		status["redis"] = "connected"
	} else {
		status["redis"] = "disconnected"
	}

	json.NewEncoder(w).Encode(status)
}

// MetricsHandler exposes Prometheus metrics
func (s *SocialMediaServer) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	// Update circuit breaker metrics before serving
	s.updateCircuitBreakerMetrics()
	// Update Redis metrics
	s.updateRedisMetrics()
	// Update rate limiter metrics
	s.updateRateLimiterMetrics()
	promhttp.Handler().ServeHTTP(w, r)
}

// rateLimitMiddleware creates a middleware that rate limits requests
func (s *SocialMediaServer) rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.rateLimitConfig.Enabled {
			next(w, r)
			return
		}

		if s.rateLimiter.Allow() {
			next(w, r)
			requestsAllowed.WithLabelValues(r.URL.Path).Inc()
			return
		}

		// Rate limit exceeded
		requestsBlocked.WithLabelValues(r.URL.Path).Inc()
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
	}
}

// metricsMiddleware wraps handlers to collect metrics
func (s *SocialMediaServer) metricsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path
		method := r.Method

		// Call the next handler
		next(w, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		httpRequestsTotal.WithLabelValues(path, method, "200").Inc()
		httpRequestDuration.WithLabelValues(path, method).Observe(duration)
	}
}

// PostToSocialMedia handles posting to social media platforms
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

	// Get video info with caching
	videoInfo := s.getVideoInfoWithCache(request.VideoID)
	if videoInfo.ID == 0 {
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}

	// Generate hashtags based on video content
	hashtags := s.generateHashtags(videoInfo.Title)

	// Post to the selected platform with resilience patterns
	var platformID string
	var err error

	start := time.Now()
	switch request.Platform {
	case "instagram":
		platformID, err = s.postToInstagramWithResilience(videoInfo.FilePath, request.Caption, hashtags)
	case "youtube":
		platformID, err = s.uploadToYouTubeWithResilience(videoInfo.FilePath, request.Caption, hashtags)
	case "tiktok":
		platformID, err = s.uploadToTikTokWithResilience(videoInfo.FilePath, request.Caption, hashtags)
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

	// Record metrics
	socialPostsTotal.WithLabelValues(request.Platform, "success").Inc()
	socialPostDuration.WithLabelValues(request.Platform).Observe(time.Since(start).Seconds())
}

// SchedulePost handles scheduling posts for later
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
	videoInfo := s.getVideoInfoWithCache(request.VideoID)
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
	})
}

// GetAnalytics returns analytics data for social media posts
func (s *SocialMediaServer) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Try to get cached analytics first
	cachedAnalytics, err := s.getCachedAnalytics()
	if err == nil && cachedAnalytics != nil {
		// Return cached data
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cachedAnalytics)
		return
	}

	// In a real app, this would query the database for analytics
	// For MVP, we'll return mock data
	analyticsData := map[string]interface{}{
		"total_posts":      42,
		"total_engagement": 1250,
		"platform_breakdown": map[string]int{
			"instagram": 15,
			"youtube":   18,
			"tiktok":    9,
		},
		"engagement_rate": 0.08,
		"top_performing_posts": []map[string]interface{}{
			{
				"id":          1,
				"platform":    "instagram",
				"engagement":  150,
				"caption":     "Amazing story!",
				"posted_at":   time.Now().AddDate(0, 0, -2).Unix(),
			},
			{
				"id":          2,
				"platform":    "youtube",
				"engagement":  300,
				"caption":     "Amazing story!",
				"posted_at":   time.Now().AddDate(0, 0, -1).Unix(),
			},
		},
	}

	// Cache the analytics data for 5 minutes
	if s.redisClient != nil {
		ctx := context.Background()
		jsonData, _ := json.Marshal(analyticsData)
		s.redisClient.Set(ctx, "analytics:data", jsonData, 5*time.Minute)
		cacheOperationsTotal.WithLabelValues("set", "analytics").Inc()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analyticsData)
}

// GetPosts returns a list of social media posts
func (s *SocialMediaServer) GetPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Try to get cached posts first
	cachedPosts, err := s.getCachedPosts()
	if err == nil && cachedPosts != nil {
		// Return cached data
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cachedPosts)
		return
	}

	// In a real app, this would query the database with pagination
	// For MVP, we'll return mock data
	posts := []SocialMediaPost{
		{
			VideoID:         1,
			Platform:        "instagram",
			PlatformID:      "ig_123456789",
			Caption:         "Check out this amazing story!",
			Hashtags:        []string{"#amazing", "#story", "#instagram"},
			PostedAt:        time.Now().AddDate(0, 0, -1),
			EngagementMetrics: map[string]interface{}{
				"likes":     24,
				"comments":  5,
				"shares":    3,
				"saves":     8,
				"views":     0,
			},
			Status:  "posted",
		},
		{
			VideoID:         2,
			Platform:        "youtube",
			PlatformID:      "yt_987654321",
			Caption:         "Check out this amazing video!",
			Hashtags:        []string{"#amazing", "#video", "#youtube"},
			PostedAt:        time.Now().AddDate(0, 0, -2),
			EngagementMetrics: map[string]interface{}{
				"likes":     150,
				"comments":  25,
				"shares":    12,
				"saves":     0,
				"views":     1500,
			},
			PlatformSpecific: map[string]string{
				"video_url": "https://youtube.com/watch?v=yt_987654321",
			},
			Status:  "posted",
		},
	}

	// Cache the posts data for 2 minutes
	if s.redisClient != nil {
		ctx := context.Background()
		jsonData, _ := json.Marshal(posts)
		s.redisClient.Set(ctx, "posts:data", jsonData, 2*time.Minute)
		cacheOperationsTotal.WithLabelValues("set", "posts").Inc()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// GetSupportedPlatforms returns the list of supported social media platforms
func (s *SocialMediaServer) GetSupportedPlatforms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]string{"instagram", "youtube", "tiktok"})
}

// Helper functions
func boolToString(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}

// isSupportedPlatform checks if a platform is supported
func (s *SocialMediaServer) isSupportedPlatform(platform string) bool {
	switch platform {
	case "instagram", "youtube", "tiktok":
		return true
	default:
		return false
	}
}

// getVideoInfoWithCache retrieves video information with Redis caching
func (s *SocialMediaServer) getVideoInfoWithCache(videoID int) Video {
	// Try to get from cache first
	if s.redisClient != nil {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("video:info:%d", videoID)

		cachedData, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil && cachedData != "" {
			var video Video
			if err := json.Unmarshal([]byte(cachedData), &video); err == nil {
				cacheOperationsTotal.WithLabelValues("get", "video_hit").Inc()
				return video
			}
		}
		cacheOperationsTotal.WithLabelValues("get", "video_miss").Inc()
	}

	// Fallback to original implementation
	video := s.getVideoInfo(videoID)

	// Cache the result for 10 minutes
	if s.redisClient != nil && video.ID != 0 {
		ctx := context.Background()
		cacheKey := fmt.Sprintf("video:info:%d", videoID)
		jsonData, _ := json.Marshal(video)
		s.redisClient.Set(ctx, cacheKey, jsonData, 10*time.Minute)
		cacheOperationsTotal.WithLabelValues("set", "video").Inc()
	}

	return video
}

// getVideoInfo retrieves video information (mock implementation)
// In a real app, this would call the content-manager service
func (s *SocialMediaServer) getVideoInfo(videoID int) Video {
	// Mock video data
	videos := map[int]Video{
		1: {ID: 1, Title: "The Tortoise and the Hare", FilePath: "/app/videos/tortoise_and_hare.mp4"},
		2: {ID: 2, Title: "The Sun and the Moon", FilePath: "/app/videos/sun_and_moon.mp4"},
		3: {ID: 3, Title: "The Clever Rabbit and the Lion", FilePath: "/app/videos/clever_rabbit_lion.mp4"},
		4: {ID: 4, Title: "The Wise Old Owl", FilePath: "/app/videos/wise_old_owl.mp4"},
	}
	if video, exists := videos[videoID]; exists {
		return video
	}
	return Video{}
}

// generateHashtags generates hashtags based on video title
func (s *SocialMediaServer) generateHashtags(title string) []string {
	// Simple hashtag generation based on title words
	words := strings.Fields(strings.ToLower(title))
	hashtags := []string{"#ShortStory", "#StoryTime"}

	for _, word := range words {
		if len(word) > 3 { // Only use words longer than 3 characters
			hashtags = append(hashtags, "#"+strings.Title(word))
		}
	}

	// Limit to 5 hashtags total
	if len(hashtags) > 5 {
		hashtags = hashtags[:5]
	}

	return hashtags
}

// getCachedAnalytics retrieves cached analytics data
func (s *SocialMediaServer) getCachedAnalytics() (map[string]interface{}, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	ctx := context.Background()
	cachedData, err := s.redisClient.Get(ctx, "analytics:data").Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(cachedData), &data); err != nil {
		return nil, err
	}

	cacheOperationsTotal.WithLabelValues("get", "analytics_hit").Inc()
	return data, nil
}

// getCachedPosts retrieves cached posts data
func (s *SocialMediaServer) getCachedPosts() ([]SocialMediaPost, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	ctx := context.Background()
	cachedData, err := s.redisClient.Get(ctx, "posts:data").Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, err
	}

	var posts []SocialMediaPost
	if err := json.Unmarshal([]byte(cachedData), &posts); err != nil {
		return nil, err
	}

	cacheOperationsTotal.WithLabelValues("get", "posts_hit").Inc()
	return posts, nil
}

// postToInstagramWithResilience posts to Instagram with circuit breaker and retry logic
func (s *SocialMediaServer) postToInstagramWithResilience(videoPath, caption string, hashtags []string) (string, error) {
	// Check if Instagram is enabled
	if !s.instagramConfig.Enabled {
		return s.simulateInstagramPost(videoPath, caption, hashtags)
	}

	// Use circuit breaker
	result, err := s.instagramCB.Execute(func() (interface{}, error) {
		return s.postToInstagramWithRetry(videoPath, caption, hashtags)
	})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

// postToInstagramWithRetry attempts to post to Instagram with retry logic
func (s *SocialMediaServer) postToInstagramWithRetry(videoPath, caption string, hashtags []string) (string, error) {
	var err error
	var result string

	// Try with exponential backoff
	for attempt := 0; attempt <= s.instagramConfig.MaxRetries; attempt++ {
		result, err = s.postToInstagram(videoPath, caption, hashtags)
		if err == nil {
			// Success, break out of retry loop
			break
		}

		// If we've exhausted retries, return the error
		if attempt == s.instagramConfig.MaxRetries {
			return "", fmt.Errorf("failed after %d attempts: %w", attempt+1, err)
		}

		// Calculate delay with exponential backoff and jitter
		delay := s.retryWithBackoff(attempt)
		time.Sleep(delay)
	}

	return result, err
}

// postToInstagram actually posts to Instagram (simulated)
func (s *SocialMediaServer) postToInstagram(videoPath, caption string, hashtags []string) (string, error) {
	// In a real implementation, this would make an actual API call to Instagram
	// For now, we'll simulate it
	return s.simulateInstagramPost(videoPath, caption, hashtags)
}

// simulateInstagramPost simulates posting to Instagram
func (s *SocialMediaServer) simulateInstagramPost(videoPath, caption string, hashtags []string) (string, error) {
	// Simulate network delay
	time.Sleep(time.Duration(rand.Intn(1000)+500) * time.Millisecond)

	// Simulate occasional failure (10% failure rate)
	if rand.Intn(10) == 0 {
		return "", fmt.Errorf("simulated Instagram API error")
	}

	// Generate a fake Instagram ID
	timestamp := time.Now().Unix()
	igID := fmt.Sprintf("ig_%d", timestamp)

	// Log the simulated post
	log.Printf("Simulating Instagram post:")
	log.Printf("  Video: %s", videoPath)
	log.Printf("  Caption: %s", caption)
	log.Printf("  Hashtags: %v", hashtags)
	log.Printf("  Instagram ID: %s", igID)

	return igID, nil
}

// uploadToYouTubeWithResilience uploads to YouTube with circuit breaker and retry logic
func (s *SocialMediaServer) uploadToYouTubeWithResilience(videoPath, caption string, hashtags []string) (string, error) {
	// Check if YouTube is enabled
	if !s.youtubeConfig.Enabled {
		return s.simulateYouTubeUpload(videoPath, caption, hashtags)
	}

	// Use circuit breaker
	result, err := s.youtubeCB.Execute(func() (interface{}, error) {
		return s.uploadToYouTubeWithRetry(videoPath, caption, hashtags)
	})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

// uploadToYouTubeWithRetry attempts to upload to YouTube with retry logic
func (s *SocialMediaServer) uploadToYouTubeWithRetry(videoPath, caption string, hashtags []string) (string, error) {
	var err error
	var result string

	// Try with exponential backoff
	for attempt := 0; attempt <= s.youtubeConfig.MaxRetries; attempt++ {
		result, err = s.uploadToYouTube(videoPath, caption, hashtags)
		if err == nil {
			// Success, break out of retry loop
			break
		}

		// If we've exhausted retries, return the error
		if attempt == s.youtubeConfig.MaxRetries {
			return "", fmt.Errorf("failed after %d attempts: %w", attempt+1, err)
		}

		// Calculate delay with exponential backoff and jitter
		delay := s.retryWithBackoff(attempt)
		time.Sleep(delay)
	}

	return result, err
}

// uploadToYouTube actually uploads to YouTube (simulated)
func (s *SocialMediaServer) uploadToYouTube(videoPath, caption string, hashtags []string) (string, error) {
	// In a real implementation, this would make an actual API call to YouTube
	// For now, we'll simulate it
	return s.simulateYouTubeUpload(videoPath, caption, hashtags)
}

// simulateYouTubeUpload simulates uploading to YouTube
func (s *SocialMediaServer) simulateYouTubeUpload(videoPath, caption string, hashtags []string) (string, error) {
	// Simulate network delay
	time.Sleep(time.Duration(rand.Intn(2000)+1000) * time.Millisecond)

	// Simulate occasional failure (5% failure rate)
	if rand.Intn(20) == 0 {
		return "", fmt.Errorf("simulated YouTube API error")
	}

	// Generate a fake YouTube video ID
	timestamp := time.Now().Unix()
	ytID := fmt.Sprintf("yt_%d", timestamp)

	// Log the simulated upload
	log.Printf("Simulating YouTube upload:")
	log.Printf("  Video: %s", videoPath)
	log.Printf("  Title: %s", caption)
	log.Printf("  Hashtags: %v", hashtags)
	log.Printf("  YouTube Video ID: %s", ytID)

	return ytID, nil
}

// uploadToTikTokWithResilience uploads to TikTok with circuit breaker and retry logic
func (s *SocialMediaServer) uploadToTikTokWithResilience(videoPath, caption string, hashtags []string) (string, error) {
	// Check if TikTok is enabled
	if !s.tiktokConfig.Enabled {
		return s.simulateTikTokUpload(videoPath, caption, hashtags)
	}

	// Use circuit breaker
	result, err := s.tiktokCB.Execute(func() (interface{}, error) {
		return s.uploadToTikTokWithRetry(videoPath, caption, hashtags)
	})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

// uploadToTikTokWithRetry attempts to upload to TikTok with retry logic
func (s *SocialMediaServer) uploadToTikTokWithRetry(videoPath, caption string, hashtags []string) (string, error) {
	var err error
	var result string

	// Try with exponential backoff
	for attempt := 0; attempt <= s.tiktokConfig.MaxRetries; attempt++ {
		result, err = s.uploadToTikTok(videoPath, caption, hashtags)
		if err == nil {
			// Success, break out of retry loop
			break
		}

		// If we've exhausted retries, return the error
		if attempt == s.tiktokConfig.MaxRetries {
			return "", fmt.Errorf("failed after %d attempts: %w", attempt+1, err)
		}

		// Calculate delay with exponential backoff and jitter
		delay := s.retryWithBackoff(attempt)
		time.Sleep(delay)
	}

	return result, err
}

// uploadToTikTok actually uploads to TikTok (simulated)
func (s *SocialMediaServer) uploadToTikTok(videoPath, caption string, hashtags []string) (string, error) {
	// In a real implementation, this would make an actual API call to TikTok
	// For now, we'll simulate it
	return s.simulateTikTokUpload(videoPath, caption, hashtags)
}

// simulateTikTokUpload simulates uploading to TikTok
func (s *SocialMediaServer) simulateTikTokUpload(videoPath, caption string, hashtags []string) (string, error) {
	// Simulate network delay
	time.Sleep(time.Duration(rand.Intn(2000)+1000) * time.Millisecond)

	// Simulate occasional failure (5% failure rate)
	if rand.Intn(20) == 0 {
		return "", fmt.Errorf("simulated TikTok API error")
	}

	// Generate a fake TikTok video ID
	timestamp := time.Now().Unix()
	tkID := fmt.Sprintf("tk_%d", timestamp)

	// Log the simulated upload
	log.Printf("Simulating TikTok upload:")
	log.Printf("  Video: %s", videoPath)
	log.Printf("  Caption: %s", caption)
	log.Printf("  Hashtags: %v", hashtags)
	log.Printf("  TikTok Video ID: %s", tkID)

	return tkID, nil
}

// retryWithBackoff calculates delay with exponential backoff and jitter
func (s *SocialMediaServer) retryWithBackoff(attempt int) time.Duration {
	// Exponential backoff: 2^attempt * 100ms, with jitter
	base := time.Duration(100) * time.Millisecond * time.Duration(1<<uint(attempt))
	jitter := time.Duration(rand.Intn(int(base))) // 0 to base milliseconds
	return base + jitter/2 // base + up to 50% jitter
}

// updateCircuitBreakerMetrics updates Prometheus metrics for circuit breakers
func (s *SocialMediaServer) updateCircuitBreakerMetrics() {
	// Update Instagram circuit breaker
	s.instagramCB.mu.RLock()
	switch s.instagramCB.state {
	case Open:
		cbStateGauge.WithLabelValues("instagram").Set(1)
	case HalfOpen:
		cbStateGauge.WithLabelValues("instagram").Set(2)
	default: // Closed
		cbStateGauge.WithLabelValues("instagram").Set(0)
	}
	s.instagramCB.mu.RUnlock()

	// Update YouTube circuit breaker
	s.youtubeCB.mu.RLock()
	switch s.youtubeCB.state {
	case Open:
		cbStateGauge.WithLabelValues("youtube").Set(1)
	case HalfOpen:
		cbStateGauge.WithLabelValues("youtube").Set(2)
	default: // Closed
		cbStateGauge.WithLabelValues("youtube").Set(0)
	}
	s.youtubeCB.mu.RUnlock()

	// Update TikTok circuit breaker
	s.tiktokCB.mu.RLock()
	switch s.tiktokCB.state {
	case Open:
		cbStateGauge.WithLabelValues("tiktok").Set(1)
	case HalfOpen:
		cbStateGauge.WithLabelValues("tiktok").Set(2)
	default: // Closed
		cbStateGauge.WithLabelValues("tiktok").Set(0)
	}
	s.tiktokCB.mu.RUnlock()
}

// updateRedisMetrics updates Prometheus metrics for Redis
func (s *SocialMediaServer) updateRedisMetrics() {
	if s.redisClient != nil {
		ctx := context.Background()
		info, err := s.redisClient.Info(ctx).Result()
		if err == nil {
			// Parse Redis info to extract metrics
			// For simplicity, we'll just set a gauge indicating Redis is connected
			redisConnectedGauge.Set(1)

			// Extract some basic info if available
			lines := strings.Split(info, "\r\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "connected_clients:") {
					parts := strings.Split(line, ":")
					if len(parts) == 2 {
						if clients, err := strconv.Atoi(parts[1]); err == nil {
							redisClientsGauge.Set(float64(clients))
						}
					}
				} else if strings.HasPrefix(line, "used_memory:") {
					parts := strings.Split(line, ":")
					if len(parts) == 2 {
						if mem, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
							redisMemoryGauge.Set(float64(mem))
						}
					}
				}
			}
		} else {
			redisConnectedGauge.Set(0)
		}
	} else {
		redisConnectedGauge.Set(0)
	}
}

// updateRateLimiterMetrics updates Prometheus metrics for rate limiter
func (s *SocialMediaServer) updateRateLimiterMetrics() {
	// Update rate limiter configuration metrics
	rateLimitRPSGauge.Set(s.rateLimitConfig.RequestsPerSecond)
	rateLimitBurstGauge.Set(float64(s.rateLimitConfig.Burst))
	rateLimitEnabledGauge.Set(boolToFloat(s.rateLimitConfig.Enabled))
}

// Helper function to convert bool to float64 for Prometheus
func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

// Initialize Prometheus metrics
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method", "status_code"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Duration of HTTP requests",
			Buckets: prometheus.ExponentialBuckets(0.005, 2, 10),
		},
		[]string{"path", "method"},
	)
	socialPostsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "social_media_posts_total",
			Help: "Total number of social media posts",
		},
		[]string{"platform", "status"},
	)
	socialPostDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "social_media_post_duration_seconds",
			Help:    "Duration of social media post operations",
			Buckets: prometheus.ExponentialBuckets(0.005, 2, 10),
		},
		[]string{"platform"},
	)
	cbStateGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Current state of circuit breakers (0=closed, 1=open, 2=half-open)",
		},
		[]string{"service"},
	)
	// Redis metrics
	redisConnectedGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "redis_connected",
			Help: "Redis connection status (0=disconnected, 1=connected)",
		},
	)
	redisClientsGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "redis_clients_connected",
			Help: "Number of client connections to Redis",
		},
	)
	redisMemoryGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "redis_used_memory_bytes",
			Help: "Used memory by Redis in bytes",
		},
	)
	// Rate limiter metrics
	rateLimitRPSGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "rate_limit_requests_per_second",
			Help: "Configured requests per second limit",
		},
	)
	rateLimitBurstGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "rate_limit_burst",
			Help: "Configured burst size",
		},
	)
	rateLimitEnabledGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "rate_limit_enabled",
			Help: "Whether rate limiting is enabled (0=disabled, 1=enabled)",
		},
	)
	// Rate limiter request counters
	requestsAllowed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_requests_allowed_total",
			Help: "Total number of requests allowed by rate limiter",
		},
		[]string{"endpoint"},
	)
	requestsBlocked = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_requests_blocked_total",
			Help: "Total number of requests blocked by rate limiter",
		},
		[]string{"endpoint"},
	)
	// Cache metrics
	cacheOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_operations_total",
			Help: "Total number of cache operations",
		},
		[]string{"operation", "type"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(socialPostsTotal)
	prometheus.MustRegister(socialPostDuration)
	prometheus.MustRegister(cbStateGauge)
	prometheus.MustRegister(redisConnectedGauge)
	prometheus.MustRegister(redisClientsGauge)
	prometheus.MustRegister(redisMemoryGauge)
	prometheus.MustRegister(rateLimitRPSGauge)
	prometheus.MustRegister(rateLimitBurstGauge)
	prometheus.MustRegister(rateLimitEnabledGauge)
	prometheus.MustRegister(requestsAllowed)
	prometheus.MustRegister(requestsBlocked)
	prometheus.MustRegister(cacheOperationsTotal)
}

func main() {
	// Load environment variables from .env file
	godotenv.Load()

	// Load configuration from environment
	instagramConfig := InstagramConfig{
		AppID:       os.Getenv("INSTAGRAM_APP_ID"),
		AppSecret:   os.Getenv("INSTAGRAM_APP_SECRET"),
		AccessToken: os.Getenv("INSTAGRAM_ACCESS_TOKEN"),
		Enabled:     os.Getenv("INSTAGRAM_APP_ID") != "" && os.Getenv("INSTAGRAM_APP_SECRET") != "" && os.Getenv("INSTAGRAM_ACCESS_TOKEN") != "",
		MaxRetries:  3,
	}

	youtubeConfig := YouTubeConfig{
		APIKey:    os.Getenv("YOUTUBE_API_KEY"),
		Enabled:   os.Getenv("YOUTUBE_API_KEY") != "",
		MaxRetries: 3,
	}

	tiktokConfig := TikTokConfig{
		ClientKey:    os.Getenv("TIKTOK_CLIENT_KEY"),
		ClientSecret: os.Getenv("TIKTOK_CLIENT_SECRET"),
		AccessToken:  os.Getenv("TIKTOK_ACCESS_TOKEN"),
		Enabled:      os.Getenv("TIKTOK_CLIENT_KEY") != "" && os.Getenv("TIKTOK_CLIENT_SECRET") != "" && os.Getenv("TIKTOK_ACCESS_TOKEN") != "",
		MaxRetries:   3,
	}

	// Load rate limit configuration from environment
	rateLimitConfig := RateLimitConfig{
		RequestsPerSecond: 5.0, // Default: 5 requests per second
		Burst:             10,  // Default: burst of 10
		Enabled:           true, // Default: enabled
	}

	if rps := os.Getenv("RATE_LIMIT_RPS"); rps != "" {
		if val, err := strconv.ParseFloat(rps, 64); err == nil {
			rateLimitConfig.RequestsPerSecond = val
		}
	}

	if burst := os.Getenv("RATE_LIMIT_BURST"); burst != "" {
		if val, err := strconv.Atoi(burst); err == nil {
			rateLimitConfig.Burst = val
		}
	}

	if enabled := os.Getenv("RATE_LIMIT_ENABLED"); enabled != "" {
		if val, err := strconv.ParseBool(enabled); err == nil {
			rateLimitConfig.Enabled = val
		}
	}

	// Initialize server
	server := NewSocialMediaServer()
	server.instagramConfig = instagramConfig
	server.youtubeConfig = youtubeConfig
	server.tiktokConfig = tiktokConfig
	server.rateLimitConfig = rateLimitConfig

	// Update rate limiter with config from environment
	if rateLimitConfig.Enabled {
		server.rateLimiter = rate.NewLimiter(rate.Limit(rateLimitConfig.RequestsPerSecond), rateLimitConfig.Burst)
	} else {
		// Disable rate limiting by setting a very high limit
		server.rateLimiter = rate.NewLimiter(rate.Inf, 0)
	}

	// Initialize Redis client if configured
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr != "" {
		opt, err := redis.ParseURL(redisAddr)
		if err == nil {
			server.redisClient = redis.NewClient(opt)
			// Test connection
			_, err = server.redisClient.Ping(context.Background()).Result()
			if err != nil {
				log.Printf("Warning: Failed to connect to Redis: %v", err)
				server.redisClient = nil
			} else {
				log.Println("Connected to Redis")
			}
		} else {
			log.Printf("Warning: Invalid Redis URL: %v", err)
		}
	} else {
		log.Println("Redis not configured, caching disabled")
	}

	// Setup routes
	server.SetupRoutes()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "9002"
	}
	log.Printf("Social Media Bot server starting on port %s", port)
	log.Printf("Rate limiting: %v (%.2f req/sec, burst=%d)",
		rateLimitConfig.Enabled, rateLimitConfig.RequestsPerSecond, rateLimitConfig.Burst)
	log.Fatal(http.ListenAndServe(":"+port, server.router))
}