package models

import "time"

// Room is a live room published to the client live catalogue.
type Room struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Title          string     `gorm:"size:255;not null" json:"title"`
	StreamerName   string     `gorm:"size:64;not null" json:"streamerName"`
	StreamerAvatar string     `gorm:"size:512" json:"streamerAvatar"`
	CoverURL       string     `gorm:"size:512" json:"coverUrl"`
	Viewers        int        `gorm:"not null;default:0" json:"viewers"`
	GameID         *uint      `gorm:"index" json:"gameId"`
	GameName       string     `gorm:"size:128" json:"gameName"`
	ServerID       *uint      `gorm:"index" json:"serverId"`
	ServerName     string     `gorm:"size:128" json:"serverName"`
	Status         string     `gorm:"size:16;not null;index" json:"status"`
	RoomURL        string     `gorm:"size:512;not null" json:"roomUrl"`
	StartedAt      time.Time  `gorm:"index" json:"startedAt"`
	EndedAt        *time.Time `json:"endedAt"`
	Sort           int        `gorm:"not null;default:0" json:"sort"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

func (Room) TableName() string { return "gb_live_rooms" }
