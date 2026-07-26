package models

// Streamer is the public profile of a live streamer.
type Streamer struct {
	ID            string `gorm:"primaryKey;size:64" json:"id"`
	Name          string `gorm:"size:64;not null" json:"name"`
	AvatarURL     string `gorm:"size:512" json:"avatarUrl"`
	CoverURL      string `gorm:"size:512" json:"coverUrl"`
	Description   string `gorm:"size:512" json:"description"`
	Fans          int    `gorm:"not null;default:0;index" json:"fans"`
	Following     bool   `gorm:"not null;default:false" json:"following"`
	IsLive        bool   `gorm:"not null;default:false;index" json:"isLive"`
	CurrentRoomID string `gorm:"size:64" json:"currentRoomId"`
	Sort          int    `gorm:"not null;default:0;index" json:"sort"`
}

func (Streamer) TableName() string { return "gb_live_streamers" }
