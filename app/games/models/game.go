package models

import "time"

// Game is a publishable game shown in the C-end catalogue.
type Game struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:128;not null;index" json:"name"`
	Slug          string    `gorm:"size:128;not null;uniqueIndex" json:"slug"`
	Description   string    `gorm:"type:text" json:"description"`
	IconURL       string    `gorm:"size:512" json:"iconUrl"`
	Category      string    `gorm:"size:64;index" json:"category"`
	GameType      string    `gorm:"size:32;index" json:"gameType"`
	Publisher     string    `gorm:"size:128" json:"publisher"`
	Rating        float64   `gorm:"type:decimal(3,1);default:0" json:"rating"`
	DownloadCount int64     `gorm:"not null;default:0" json:"downloadCount"`
	VersionTags   string    `gorm:"size:512" json:"versionTags"`
	Status        string    `gorm:"size:16;not null;default:draft;index" json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (Game) TableName() string { return "games" }
