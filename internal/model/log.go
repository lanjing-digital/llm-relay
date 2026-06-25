package model

import "time"

type Log struct {
	ID              uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt       time.Time `json:"created_at"`
	Method          string    `json:"method" gorm:"not null"`
	Path            string    `json:"path" gorm:"not null"`
	ClientIP        string    `json:"client_ip" gorm:"not null"`
	ExternalModel   string    `json:"external_model"`
	TargetModel     string    `json:"target_model"`
	UpstreamURL     string    `json:"upstream_url"`
	StatusCode      int       `json:"status_code" gorm:"not null"`
	DurationMs      int64     `json:"duration_ms" gorm:"not null"`
	RequestSnippet  string    `json:"request_snippet" gorm:"type:text"`
	ResponseSnippet string    `json:"response_snippet" gorm:"type:text"`
	Error           string    `json:"error" gorm:"type:text"`
}

