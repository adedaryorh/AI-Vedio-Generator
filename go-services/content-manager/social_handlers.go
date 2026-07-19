package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var supportedPlatforms = map[string]bool{"instagram": true, "youtube": true, "tiktok": true}

type socialPostRequest struct {
	VideoID          uint              `json:"video_id" binding:"required"`
	Platform         string            `json:"platform" binding:"required"`
	Caption          string            `json:"caption"`
	Hashtags         []string          `json:"hashtags"`
	PlatformSpecific map[string]string `json:"platform_specific"`
	ScheduleAt       string            `json:"schedule_at"`
}
type socialPostResponse struct {
	ID                                    uint
	VideoID                               uint
	Platform, PlatformID, Caption, Status string
	Hashtags                              []string
	ScheduledAt, PostedAt                 *time.Time
	CreatedAt                             time.Time
	EngagementMetrics                     map[string]interface{}
	PlatformSpecific                      map[string]string
}

func (s *Server) PublishSocialPost(c *gin.Context) {
	var req socialPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	post, err := s.publish(req)
	if err != nil {
		c.JSON(statusForSocialError(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toSocialResponse(*post))
}
func (s *Server) publish(req socialPostRequest) (*SocialPost, error) {
	platform := strings.ToLower(strings.TrimSpace(req.Platform))
	if !supportedPlatforms[platform] {
		return nil, errors.New("unsupported platform")
	}
	var video Video
	if err := s.DB.First(&video, req.VideoID).Error; err != nil {
		return nil, errors.New("video not found")
	}
	if video.Status != StatusReady {
		return nil, errors.New("video is not ready for publishing")
	}
	if len(req.Hashtags) == 0 {
		req.Hashtags = generateHashtags(s.videoTitle(video))
	}
	start := time.Now()
	result, err := s.CircuitBreakers[platform].Execute(func() (interface{}, error) { return s.publishWithRetry(platform, video, req) })
	status := "success"
	if err != nil {
		status = "failed"
	}
	socialRequests.WithLabelValues(platform, status).Inc()
	socialDuration.WithLabelValues(platform).Observe(time.Since(start).Seconds())
	now := time.Now().UTC()
	metrics, _ := json.Marshal(emptyEngagement())
	if req.PlatformSpecific == nil {
		req.PlatformSpecific = map[string]string{}
	}
	specific, _ := json.Marshal(req.PlatformSpecific)
	post := SocialPost{VideoID: req.VideoID, Platform: platform, Caption: req.Caption, Hashtags: strings.Join(req.Hashtags, ","), Status: "posted", PostedAt: &now, EngagementMetrics: string(metrics), PlatformSpecific: string(specific)}
	if err != nil {
		post.Status = "failed"
		_ = s.DB.Create(&post).Error
		return nil, err
	}
	post.PlatformID = result.(string)
	addPlatformResult(req.PlatformSpecific, platform, post.PlatformID)
	specific, _ = json.Marshal(req.PlatformSpecific)
	post.PlatformSpecific = string(specific)
	if err := s.DB.Create(&post).Error; err != nil {
		return nil, err
	}
	s.invalidatePostsCache()
	return &post, nil
}
func (s *Server) publishWithRetry(platform string, video Video, req socialPostRequest) (string, error) {
	retries := s.maxRetries(platform)
	var last error
	for attempt := 0; attempt <= retries; attempt++ {
		var id string
		switch platform {
		case "instagram":
			id, last = s.postInstagram(video, req)
		case "youtube":
			id, last = s.postYouTube(video, req)
		case "tiktok":
			id, last = s.postTikTok(video, req)
		}
		if last == nil {
			return id, nil
		}
		time.Sleep(time.Duration(attempt+1) * 250 * time.Millisecond)
	}
	return "", last
}
func (s *Server) maxRetries(platform string) int {
	switch platform {
	case "instagram":
		return s.config.Instagram.MaxRetries
	case "youtube":
		return s.config.YouTube.MaxRetries
	default:
		return s.config.TikTok.MaxRetries
	}
}
func (s *Server) postInstagram(video Video, req socialPostRequest) (string, error) {
	cfg := s.config.Instagram
	if cfg.AppID == "" || cfg.AccessToken == "" {
		return simulatedID("instagram", video.ID), nil
	}
	endpoint := fmt.Sprintf("https://graph.facebook.com/v20.0/%s/media", cfg.AppID)
	values := url.Values{"video_url": {video.FilePath}, "caption": {captionWithTags(req)}, "access_token": {cfg.AccessToken}, "media_type": {"REELS"}}
	return platformHTTPPost(endpoint, strings.NewReader(values.Encode()), "application/x-www-form-urlencoded")
}
func (s *Server) postYouTube(video Video, req socialPostRequest) (string, error) {
	cfg := s.config.YouTube
	if cfg.APIKey == "" {
		return simulatedID("youtube", video.ID), nil
	}
	payload, _ := json.Marshal(map[string]interface{}{"snippet": map[string]interface{}{"title": req.Caption, "description": captionWithTags(req), "tags": req.Hashtags}, "status": map[string]string{"privacyStatus": "private"}})
	return platformHTTPPost("https://www.googleapis.com/youtube/v3/videos?part=snippet,status&key="+url.QueryEscape(cfg.APIKey), bytes.NewReader(payload), "application/json")
}
func (s *Server) postTikTok(video Video, req socialPostRequest) (string, error) {
	cfg := s.config.TikTok
	if cfg.AccessToken == "" {
		return simulatedID("tiktok", video.ID), nil
	}
	payload, _ := json.Marshal(map[string]interface{}{"post_info": map[string]string{"title": captionWithTags(req), "privacy_level": "SELF_ONLY"}, "source_info": map[string]string{"source": "PULL_FROM_URL", "video_url": video.FilePath}})
	request, _ := http.NewRequest(http.MethodPost, "https://open.tiktokapis.com/v2/post/publish/video/init/", bytes.NewReader(payload))
	request.Header.Set("Authorization", "Bearer "+cfg.AccessToken)
	request.Header.Set("Content-Type", "application/json")
	return doPlatformRequest(request)
}
func platformHTTPPost(endpoint string, body io.Reader, contentType string) (string, error) {
	req, err := http.NewRequest(http.MethodPost, endpoint, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", contentType)
	return doPlatformRequest(req)
}
func doPlatformRequest(req *http.Request) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("platform API returned %d: %s", resp.StatusCode, string(body))
	}
	var result map[string]interface{}
	_ = json.Unmarshal(body, &result)
	for _, key := range []string{"id", "publish_id", "creation_id"} {
		if value, ok := result[key]; ok {
			return fmt.Sprint(value), nil
		}
	}
	return fmt.Sprintf("remote-%d", time.Now().Unix()), nil
}
func simulatedID(platform string, id uint) string {
	return fmt.Sprintf("%s-sim-%d-%d", platform, id, time.Now().Unix())
}
func captionWithTags(req socialPostRequest) string {
	if len(req.Hashtags) == 0 {
		return req.Caption
	}
	return strings.TrimSpace(req.Caption + " #" + strings.Join(req.Hashtags, " #"))
}

func (s *Server) ScheduleSocialPost(c *gin.Context) {
	var req socialPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	platform := strings.ToLower(req.Platform)
	if !supportedPlatforms[platform] {
		c.JSON(400, gin.H{"error": "unsupported platform"})
		return
	}
	var video Video
	if err := s.cachedVideo(req.VideoID, &video); err != nil {
		c.JSON(404, gin.H{"error": "video not found"})
		return
	}
	if len(req.Hashtags) == 0 {
		req.Hashtags = generateHashtags(s.videoTitle(video))
	}
	scheduled, err := time.Parse(time.RFC3339, req.ScheduleAt)
	if err != nil || !scheduled.After(time.Now()) {
		c.JSON(400, gin.H{"error": "schedule_at must be a future RFC3339 timestamp"})
		return
	}
	metrics, _ := json.Marshal(emptyEngagement())
	specific, _ := json.Marshal(req.PlatformSpecific)
	post := SocialPost{VideoID: req.VideoID, Platform: platform, Caption: req.Caption, Hashtags: strings.Join(req.Hashtags, ","), Status: "scheduled", ScheduledAt: &scheduled, EngagementMetrics: string(metrics), PlatformSpecific: string(specific)}
	if err := s.DB.Create(&post).Error; err != nil {
		respondError(c, err)
		return
	}
	s.invalidatePostsCache()
	c.JSON(201, toSocialResponse(post))
}
func (s *Server) GetSocialPosts(c *gin.Context) {
	cacheKey := "social:posts:" + c.Query("status")
	if s.Redis != nil {
		if cached, err := s.Redis.Get(context.Background(), cacheKey).Bytes(); err == nil {
			c.Data(200, "application/json", cached)
			return
		}
	}
	var posts []SocialPost
	query := s.DB.Order("created_at DESC")
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&posts).Error; err != nil {
		respondError(c, err)
		return
	}
	response := make([]socialPostResponse, len(posts))
	for i := range posts {
		response[i] = toSocialResponse(posts[i])
	}
	if data, err := json.Marshal(response); err == nil && s.Redis != nil {
		_ = s.Redis.Set(context.Background(), cacheKey, data, 30*time.Second).Err()
	}
	c.JSON(200, response)
}
func (s *Server) GetSocialAnalytics(c *gin.Context) {
	if s.Redis != nil {
		if cached, err := s.Redis.Get(context.Background(), "social:analytics").Bytes(); err == nil {
			cacheOperations.WithLabelValues("get", "analytics_hit").Inc()
			c.Data(200, "application/json", cached)
			return
		}
		cacheOperations.WithLabelValues("get", "analytics_miss").Inc()
	}
	var total, posted, scheduled, failed int64
	s.DB.Model(&SocialPost{}).Count(&total)
	s.DB.Model(&SocialPost{}).Where("status = ?", "posted").Count(&posted)
	s.DB.Model(&SocialPost{}).Where("status = ?", "scheduled").Count(&scheduled)
	s.DB.Model(&SocialPost{}).Where("status = ?", "failed").Count(&failed)
	var top string
	s.DB.Model(&SocialPost{}).Select("platform").Where("status = ?", "posted").Group("platform").Order("count(*) DESC").Limit(1).Scan(&top)
	response := gin.H{"total_posts": total, "posted_posts": posted, "scheduled_posts": scheduled, "failed_posts": failed, "total_views": 0, "average_engagement": 0, "top_platform": top}
	if data, err := json.Marshal(response); err == nil && s.Redis != nil {
		_ = s.Redis.Set(context.Background(), "social:analytics", data, 5*time.Minute).Err()
		cacheOperations.WithLabelValues("set", "analytics").Inc()
	}
	c.JSON(200, response)
}
func (s *Server) GetSupportedPlatforms(c *gin.Context) {
	c.JSON(200, gin.H{"platforms": []gin.H{{"name": "instagram", "enabled": s.config.Instagram.AccessToken != ""}, {"name": "youtube", "enabled": s.config.YouTube.APIKey != ""}, {"name": "tiktok", "enabled": s.config.TikTok.AccessToken != ""}}, "simulation_mode": s.config.Instagram.AccessToken == "" && s.config.YouTube.APIKey == "" && s.config.TikTok.AccessToken == "", "circuit_breakers": gin.H{"instagram": s.CircuitBreakers["instagram"].State(), "youtube": s.CircuitBreakers["youtube"].State(), "tiktok": s.CircuitBreakers["tiktok"].State()}, "redis_connected": s.Redis != nil})
}
func (s *Server) RetrySocialPost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid post id"})
		return
	}
	var post SocialPost
	if err := s.DB.First(&post, id).Error; err != nil {
		respondDatabaseLookup(c, err, "Post not found")
		return
	}
	req := socialPostRequest{VideoID: post.VideoID, Platform: post.Platform, Caption: post.Caption, Hashtags: splitCSV(post.Hashtags)}
	result, err := s.publish(req)
	if err != nil {
		c.JSON(statusForSocialError(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, toSocialResponse(*result))
}
func (s *Server) CancelSocialPost(c *gin.Context) {
	var post SocialPost
	if err := s.DB.First(&post, c.Param("id")).Error; err != nil {
		respondDatabaseLookup(c, err, "Post not found")
		return
	}
	if post.Status != "scheduled" {
		c.JSON(409, gin.H{"error": "only scheduled posts can be cancelled"})
		return
	}
	if err := s.DB.Delete(&post).Error; err != nil {
		respondError(c, err)
		return
	}
	s.invalidatePostsCache()
	c.Status(204)
}
func toSocialResponse(post SocialPost) socialPostResponse {
	metrics := emptyEngagement()
	specific := map[string]string{}
	_ = json.Unmarshal([]byte(post.EngagementMetrics), &metrics)
	_ = json.Unmarshal([]byte(post.PlatformSpecific), &specific)
	return socialPostResponse{ID: post.ID, VideoID: post.VideoID, Platform: post.Platform, PlatformID: post.PlatformID, Caption: post.Caption, Hashtags: splitCSV(post.Hashtags), Status: post.Status, ScheduledAt: post.ScheduledAt, PostedAt: post.PostedAt, CreatedAt: post.CreatedAt, EngagementMetrics: metrics, PlatformSpecific: specific}
}
func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	return strings.Split(value, ",")
}
func emptyEngagement() map[string]interface{} {
	return map[string]interface{}{"likes": 0, "comments": 0, "shares": 0, "views": 0, "saves": 0}
}
func (s *Server) invalidatePostsCache() {
	if s.Redis != nil {
		keys, _ := s.Redis.Keys(context.Background(), "social:posts:*").Result()
		if len(keys) > 0 {
			_ = s.Redis.Del(context.Background(), keys...).Err()
		}
		_ = s.Redis.Del(context.Background(), "social:analytics").Err()
	}
}

func (s *Server) cachedVideo(id uint, video *Video) error {
	key := fmt.Sprintf("social:video:%d", id)
	if s.Redis != nil {
		if data, err := s.Redis.Get(context.Background(), key).Bytes(); err == nil && json.Unmarshal(data, video) == nil {
			cacheOperations.WithLabelValues("get", "video_hit").Inc()
			return nil
		}
		cacheOperations.WithLabelValues("get", "video_miss").Inc()
	}
	if err := s.DB.First(video, id).Error; err != nil {
		return err
	}
	if data, err := json.Marshal(video); err == nil && s.Redis != nil {
		_ = s.Redis.Set(context.Background(), key, data, 10*time.Minute).Err()
		cacheOperations.WithLabelValues("set", "video").Inc()
	}
	return nil
}

func (s *Server) videoTitle(video Video) string {
	var story Story
	if video.StoryID != 0 && s.DB.First(&story, video.StoryID).Error == nil && story.Title != "" {
		return story.Title
	}
	return "AI video story"
}

func generateHashtags(title string) []string {
	tags := []string{"ShortStory", "StoryTime"}
	seen := map[string]bool{"shortstory": true, "storytime": true}
	for _, word := range strings.Fields(strings.ToLower(title)) {
		word = strings.Trim(word, ".,!?;:'\"()[]{}-_/")
		if len(word) < 4 || seen[word] {
			continue
		}
		seen[word] = true
		tags = append(tags, strings.ToUpper(word[:1])+word[1:])
		if len(tags) == 7 {
			break
		}
	}
	return tags
}

func addPlatformResult(values map[string]string, platform, id string) {
	switch platform {
	case "youtube":
		values["video_url"] = "https://youtube.com/watch?v=" + id
	case "tiktok":
		values["video_url"] = "https://tiktok.com/@user/video/" + id
	case "instagram":
		values["media_id"] = id
	}
}
func statusForSocialError(err error) int {
	switch err.Error() {
	case "video not found":
		return 404
	case "video is not ready for publishing":
		return 409
	case "unsupported platform":
		return 400
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 404
	}
	return 502
}
