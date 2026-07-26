package models

import "time"

type StreamerFollow struct {
	ID         uint      `gorm:"primaryKey" json:"-"`
	UserID     uint      `gorm:"uniqueIndex:ux_live_follow;not null" json:"-"`
	StreamerID string    `gorm:"uniqueIndex:ux_live_follow;size:64;not null" json:"-"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (StreamerFollow) TableName() string { return "gb_live_streamer_follows" }
