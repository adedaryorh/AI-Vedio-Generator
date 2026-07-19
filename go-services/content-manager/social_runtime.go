package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

type RateLimitConfig struct {
	RequestsPerSecond float64
	Burst             int
	Enabled           bool
}
type InstagramConfig struct {
	AppID, AppSecret, AccessToken string
	MaxRetries                    int
}
type YouTubeConfig struct {
	APIKey, ChannelID string
	MaxRetries        int
}
type TikTokConfig struct {
	ClientKey, ClientSecret, AccessToken string
	MaxRetries                           int
}

type CircuitBreakerState string

const (
	Closed   CircuitBreakerState = "closed"
	Open     CircuitBreakerState = "open"
	HalfOpen CircuitBreakerState = "half_open"
)

type CircuitBreaker struct {
	mu                        sync.RWMutex
	state                     CircuitBreakerState
	failureCount, maxFailures int
	timeout                   time.Duration
	lastFailureTime           time.Time
}

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{state: Closed, maxFailures: maxFailures, timeout: timeout}
}
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	cb.mu.Lock()
	if cb.state == Open {
		if time.Since(cb.lastFailureTime) <= cb.timeout {
			cb.mu.Unlock()
			return nil, errors.New("circuit breaker is open")
		}
		cb.state = HalfOpen
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

var socialRequests = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "content_manager_social_requests_total", Help: "Social publishing requests."}, []string{"platform", "status"})
var socialDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "content_manager_social_request_duration_seconds", Help: "Social publishing latency."}, []string{"platform"})
var circuitBreakerState = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "content_manager_circuit_breaker_state", Help: "Circuit state: closed=0, open=1, half-open=2."}, []string{"platform"})
var redisConnected = prometheus.NewGauge(prometheus.GaugeOpts{Name: "content_manager_redis_connected", Help: "Whether Redis is connected."})
var redisClients = prometheus.NewGauge(prometheus.GaugeOpts{Name: "content_manager_redis_clients", Help: "Connected Redis clients."})
var redisMemory = prometheus.NewGauge(prometheus.GaugeOpts{Name: "content_manager_redis_memory_bytes", Help: "Redis memory use."})
var rateLimitConfigGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "content_manager_rate_limit_config", Help: "Configured rate limit values."}, []string{"setting"})
var rateLimitRequests = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "content_manager_rate_limit_requests_total", Help: "Allowed and blocked requests."}, []string{"result", "endpoint"})
var cacheOperations = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "content_manager_cache_operations_total", Help: "Redis cache operations."}, []string{"operation", "type"})

func init() {
	prometheus.MustRegister(socialRequests, socialDuration, circuitBreakerState, redisConnected, redisClients, redisMemory, rateLimitConfigGauge, rateLimitRequests, cacheOperations)
}

func (s *Server) connectRedis() {
	options, err := redis.ParseURL(s.config.RedisURL)
	if err != nil {
		return
	}
	client := redis.NewClient(options)
	if client.Ping(context.Background()).Err() == nil {
		s.Redis = client
	} else {
		_ = client.Close()
	}
}
func (s *Server) socialRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.config.RateLimit.Enabled && !s.RateLimiter.Allow() {
			rateLimitRequests.WithLabelValues("blocked", c.FullPath()).Inc()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded; retry shortly"})
			return
		}
		rateLimitRequests.WithLabelValues("allowed", c.FullPath()).Inc()
		c.Next()
	}
}
func (s *Server) MetricsHandler(c *gin.Context) {
	for name, breaker := range s.CircuitBreakers {
		value := float64(0)
		if breaker.State() == Open {
			value = 1
		}
		if breaker.State() == HalfOpen {
			value = 2
		}
		circuitBreakerState.WithLabelValues(name).Set(value)
	}
	rateLimitConfigGauge.WithLabelValues("requests_per_second").Set(s.config.RateLimit.RequestsPerSecond)
	rateLimitConfigGauge.WithLabelValues("burst").Set(float64(s.config.RateLimit.Burst))
	if s.config.RateLimit.Enabled {
		rateLimitConfigGauge.WithLabelValues("enabled").Set(1)
	} else {
		rateLimitConfigGauge.WithLabelValues("enabled").Set(0)
	}
	if s.Redis == nil {
		redisConnected.Set(0)
	} else if info, err := s.Redis.Info(context.Background()).Result(); err == nil {
		redisConnected.Set(1)
		for _, line := range strings.Split(info, "\r\n") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			value, _ := strconv.ParseFloat(parts[1], 64)
			if parts[0] == "connected_clients" {
				redisClients.Set(value)
			}
			if parts[0] == "used_memory" {
				redisMemory.Set(value)
			}
		}
	} else {
		redisConnected.Set(0)
	}
	promhttp.Handler().ServeHTTP(c.Writer, c.Request)
}
