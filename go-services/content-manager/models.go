package main

import "time"

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusReady      = "ready"
)

type Video struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	StoryID    uint      `gorm:"column:story_id" json:"story_id"`
	FilePath   string    `gorm:"column:file_path" json:"file_path"`
	Duration   int       `gorm:"column:duration" json:"duration"`
	Format     string    `gorm:"column:format" json:"format"`
	Resolution string    `gorm:"column:resolution" json:"resolution"`
	Status     string    `gorm:"column:status" json:"status"`
	Metadata   string    `gorm:"column:metadata" json:"metadata"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

type Story struct {
	ID        uint      `gorm:"primaryKey"`
	Title     string    `gorm:"column:title"`
	Content   string    `gorm:"column:content"`
	Source    string    `gorm:"column:source"`
	Category  string    `gorm:"column:category"`
	Culture   string    `gorm:"column:culture"`
	Tags      string    `gorm:"column:tags"`
	CreatedAt time.Time `gorm:"column:created_at"`
	Processed bool      `gorm:"column:processed"`
}

type InstagramPost struct {
	ID                uint      `gorm:"primaryKey"`
	VideoID           uint      `gorm:"column:video_id"`
	InstagramID       string    `gorm:"column:instagram_id"`
	Caption           string    `gorm:"column:caption"`
	Hashtags          string    `gorm:"column:hashtags"`
	PostedAt          time.Time `gorm:"column:posted_at"`
	EngagementMetrics string    `gorm:"column:engagement_metrics"`
	Status            string    `gorm:"column:status"`
}

// SocialPost stores publishing activity for every supported platform.
// The field names intentionally preserve the existing frontend API contract.
type SocialPost struct {
	ID                uint       `gorm:"primaryKey"`
	VideoID           uint       `gorm:"column:video_id;index"`
	Platform          string     `gorm:"column:platform;size:32"`
	PlatformID        string     `gorm:"column:platform_id;size:128"`
	Caption           string     `gorm:"column:caption;type:text"`
	Hashtags          string     `gorm:"column:hashtags;type:text"`
	Status            string     `gorm:"column:status;size:32;index"`
	ScheduledAt       *time.Time `gorm:"column:scheduled_at"`
	PostedAt          *time.Time `gorm:"column:posted_at"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	EngagementMetrics string     `gorm:"column:engagement_metrics;type:jsonb;default:'{}'"`
	PlatformSpecific  string     `gorm:"column:platform_specific;type:jsonb;default:'{}'"`
}

type Stats struct {
	TotalStories     int64 `json:"total_stories"`
	ProcessedStories int64 `json:"processed_stories"`
	TotalVideos      int64 `json:"total_videos"`
	PendingVideos    int64 `json:"pending_videos"`
	ReadyVideos      int64 `json:"ready_videos"`
	PostedVideos     int64 `json:"posted_videos"`
}
