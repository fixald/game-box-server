package models

import "time"

type Server struct {
	ID                   uint       `gorm:"primaryKey" json:"id"`
	GameID               uint       `gorm:"index;not null" json:"gameId"`
	Name                 string     `gorm:"size:128;not null" json:"name"`
	ImageURL             string     `gorm:"size:512" json:"imageUrl"`
	OpenTime             time.Time  `json:"openTime"`
	Status               string     `gorm:"size:24;not null;default:preview;index" json:"status"`
	MergeTime            *time.Time `json:"mergeTime"`
	MinClientVersion     string     `gorm:"size:64" json:"minClientVersion"`
	IsRecommended        bool       `json:"isRecommended"`
	RecommendationWeight int        `gorm:"not null;default:0;index" json:"recommendationWeight"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

func (Server) TableName() string { return "servers" }
