package models

import (
	"time"
	"gorm.io/gorm"
)

type Banner struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Title     string     `gorm:"size:128;not null" json:"title"`
	ImageURL  string     `gorm:"size:512;not null" json:"imageUrl"`
	LinkType  string     `gorm:"size:16;not null;default:none" json:"linkType"`
	LinkValue string     `gorm:"size:512" json:"linkValue"`
	Position  string     `gorm:"size:64;not null;index" json:"position"`
	Weight    int        `gorm:"not null;default:0" json:"weight"`
	GameID    *uint      `gorm:"index" json:"gameId"`
	StartAt   time.Time  `gorm:"index;not null" json:"startAt"`
	EndAt     time.Time  `gorm:"index;not null" json:"endAt"`
	Status    string     `gorm:"size:16;not null;default:draft;index" json:"status"`
	Sort      int        `gorm:"not null;default:0" json:"sort"`
	CreatedBy uint       `json:"createdBy"`
	UpdatedBy uint       `json:"updatedBy"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Banner) TableName() string { return "banners" }
