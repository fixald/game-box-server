package models

import "time"

// Room is a live room published to the client live catalogue.
type Room struct {
	ID              uint            `gorm:"primaryKey" json:"id"`
	Title           string          `gorm:"size:255;not null" json:"title"`
	Announcement    string          `gorm:"size:512" json:"announcement"`
	StreamerID      string          `gorm:"size:64" json:"streamerId"`
	StreamerName    string          `gorm:"size:64;not null" json:"streamerName"`
	StreamerAvatar  string          `gorm:"size:512" json:"streamerAvatar"`
	StreamerFans    int             `gorm:"not null;default:0" json:"streamerFans"`
	CoverURL        string          `gorm:"size:512" json:"coverUrl"`
	CategoryID      string          `gorm:"size:64;index" json:"categoryId"`
	CategoryName    string          `gorm:"size:64" json:"categoryName"`
	CategoryType    string          `gorm:"size:32" json:"categoryType"`
	Recommendation  bool            `gorm:"not null;default:false;index" json:"recommendation"`
	Viewers         int             `gorm:"not null;default:0" json:"viewers"`
	GameID          *uint           `gorm:"index" json:"gameId"`
	GameName        string          `gorm:"size:128" json:"gameName"`
	ServerID        *uint           `gorm:"index" json:"serverId"`
	ServerName      string          `gorm:"size:128" json:"serverName"`
	ServerStatus    string          `gorm:"size:32" json:"serverStatus"`
	Status          string          `gorm:"size:16;not null;index" json:"status"`
	RoomURL         string          `gorm:"size:512;not null" json:"roomUrl"`
	StreamProtocol  string          `gorm:"size:16;not null;default:hls" json:"streamProtocol"`
	StreamExpiresAt *time.Time      `json:"streamExpiresAt"`
	StreamQualities []StreamQuality `gorm:"serializer:json" json:"streamQualities"`
	StartedAt       time.Time       `gorm:"index" json:"startedAt"`
	EndedAt         *time.Time      `json:"endedAt"`
	Sort            int             `gorm:"not null;default:0" json:"sort"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

type StreamQuality struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (Room) TableName() string { return "gb_live_rooms" }
