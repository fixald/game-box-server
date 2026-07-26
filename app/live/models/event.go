package models

import "time"

type ClientEvent struct {
	ID           uint      `gorm:"primaryKey" json:"-"`
	UserID       uint      `gorm:"index;not null" json:"-"`
	EventType    string    `gorm:"size:32;not null;index" json:"eventType"`
	ResourceType string    `gorm:"size:32;not null;index" json:"resourceType"`
	ResourceID   string    `gorm:"size:64;not null;index" json:"resourceId"`
	Source       string    `gorm:"size:64" json:"source"`
	SessionID    string    `gorm:"size:128;index" json:"sessionId"`
	OccurredAt   time.Time `gorm:"index;not null" json:"occurredAt"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (ClientEvent) TableName() string { return "gb_client_events" }
