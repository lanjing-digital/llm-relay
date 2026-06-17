package model

import "time"

type Config struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name          string    `json:"name" gorm:"not null"`
	ExternalModel string    `json:"external_model" gorm:"not null;uniqueIndex"`
	TargetModel   string    `json:"target_model" gorm:"not null"`
	TargetBaseURL string    `json:"target_base_url" gorm:"not null"`
	TargetAPIKey  string    `json:"target_api_key" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
