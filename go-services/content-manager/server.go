package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Port        string
	DatabaseDSN string
	RabbitURL   string
	RedisURL    string
	RateLimit   RateLimitConfig
	Instagram   InstagramConfig
	YouTube     YouTubeConfig
	TikTok      TikTokConfig
}

func loadConfig() Config {
	// Local host settings take precedence; Docker Compose continues to inject .env.
	_ = godotenv.Load(".env.local", ".env")
	return Config{
		Port: valueOrDefault("PORT", "9001"),
		DatabaseDSN: fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
			valueOrDefault("DB_HOST", "postgres"), valueOrDefault("DB_USER", "storybot"),
			valueOrDefault("DB_PASSWORD", "storybotpass"), valueOrDefault("DB_NAME", "storybotdb"),
			valueOrDefault("DB_PORT", "5432")),
		RabbitURL: fmt.Sprintf("amqp://%s:%s@%s:%s/", valueOrDefault("RABBITMQ_USER", "guest"),
			valueOrDefault("RABBITMQ_PASS", "guest"), valueOrDefault("RABBITMQ_HOST", "rabbitmq"),
			valueOrDefault("RABBITMQ_PORT", "5672")),
		RedisURL:  valueOrDefault("REDIS_URL", "redis://redis:6379/0"),
		RateLimit: RateLimitConfig{RequestsPerSecond: floatEnv("RATE_LIMIT_RPS", 5), Burst: intEnv("RATE_LIMIT_BURST", 10), Enabled: boolEnv("RATE_LIMIT_ENABLED", true)},
		Instagram: InstagramConfig{AppID: os.Getenv("INSTAGRAM_APP_ID"), AppSecret: os.Getenv("INSTAGRAM_APP_SECRET"), AccessToken: os.Getenv("INSTAGRAM_ACCESS_TOKEN"), MaxRetries: intEnv("SOCIAL_MAX_RETRIES", 3)},
		YouTube:   YouTubeConfig{APIKey: os.Getenv("YOUTUBE_API_KEY"), ChannelID: os.Getenv("YOUTUBE_CHANNEL_ID"), MaxRetries: intEnv("SOCIAL_MAX_RETRIES", 3)},
		TikTok:    TikTokConfig{ClientKey: os.Getenv("TIKTOK_CLIENT_KEY"), ClientSecret: os.Getenv("TIKTOK_CLIENT_SECRET"), AccessToken: os.Getenv("TIKTOK_ACCESS_TOKEN"), MaxRetries: intEnv("SOCIAL_MAX_RETRIES", 3)},
	}
}

func intEnv(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}
func floatEnv(key string, fallback float64) float64 {
	value, err := strconv.ParseFloat(os.Getenv(key), 64)
	if err != nil {
		return fallback
	}
	return value
}
func boolEnv(key string, fallback bool) bool {
	value, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

type Server struct {
	Router          *gin.Engine
	DB              *gorm.DB
	RabbitMQ        *RabbitMQPublisher
	config          Config
	Redis           *redis.Client
	RateLimiter     *rate.Limiter
	CircuitBreakers map[string]*CircuitBreaker
}

func NewServer() (*Server, error) {
	config := loadConfig()
	db, err := gorm.Open(postgres.Open(config.DatabaseDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	if err := db.AutoMigrate(&Video{}, &Story{}, &InstagramPost{}, &SocialPost{}); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	publisher, err := NewRabbitMQPublisher(config.RabbitURL)
	if err != nil {
		log.Printf("RabbitMQ unavailable; queue events are disabled: %v", err)
	}
	router := gin.Default()
	router.Use(corsMiddleware())
	server := &Server{Router: router, DB: db, RabbitMQ: publisher, config: config, RateLimiter: rate.NewLimiter(rate.Limit(config.RateLimit.RequestsPerSecond), config.RateLimit.Burst), CircuitBreakers: map[string]*CircuitBreaker{"instagram": NewCircuitBreaker(5, 60*time.Second), "youtube": NewCircuitBreaker(5, 60*time.Second), "tiktok": NewCircuitBreaker(5, 60*time.Second)}}
	server.connectRedis()
	server.SetupRoutes()
	server.startSocialConsumer()
	return server, nil
}

func (s *Server) SetupRoutes() {
	s.Router.GET("/health", s.HealthCheck)
	s.Router.GET("/videos", s.GetVideos)
	s.Router.POST("/videos", s.CreateVideo)
	s.Router.GET("/videos/:id", s.GetVideoByID)
	s.Router.PUT("/videos/:id", s.UpdateVideo)
	s.Router.DELETE("/videos/:id", s.DeleteVideo)
	s.Router.POST("/queue", s.AddToQueue)
	s.Router.GET("/queue", s.GetQueue)
	s.Router.POST("/process", s.ProcessQueue)
	s.Router.GET("/stats", s.GetStats)
	social := s.Router.Group("/", s.socialRateLimit())
	social.POST("/post", s.PublishSocialPost)
	social.POST("/schedule", s.ScheduleSocialPost)
	social.GET("/posts", s.GetSocialPosts)
	social.GET("/analytics", s.GetSocialAnalytics)
	social.GET("/platforms", s.GetSupportedPlatforms)
	s.Router.GET("/metrics", s.MetricsHandler)
	social.POST("/posts/:id/retry", s.RetrySocialPost)
	social.DELETE("/posts/:id", s.CancelSocialPost)
}

func (s *Server) Start() error {
	log.Printf("Server starting on port %s", s.config.Port)
	return s.Router.Run(":" + s.config.Port)
}

func (s *Server) Close() {
	s.RabbitMQ.Close()
	if s.Redis != nil {
		_ = s.Redis.Close()
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", valueOrDefault("FRONTEND_URL", "http://localhost:3000"))
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
