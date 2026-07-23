package models

import "time"

type SearchItem struct {
	ID          string  `gorm:"primaryKey;size:64" json:"id"`
	Type        string  `gorm:"size:16;index;not null" json:"type"`
	Title       string  `gorm:"size:128;not null" json:"title"`
	Subtitle    string  `gorm:"size:255" json:"subtitle"`
	Description string  `gorm:"type:text" json:"description"`
	IconURL     string  `gorm:"size:512" json:"iconUrl"`
	CoverURL    string  `gorm:"size:512" json:"coverUrl"`
	Tags        string  `gorm:"size:512" json:"tags"`
	Target      string  `gorm:"type:text" json:"target"`
	Score       float64 `gorm:"default:0" json:"score"`
}

func (SearchItem) TableName() string { return "gb_search_items" }

type SearchHistory struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"-"`
	Keyword    string    `gorm:"size:128;not null" json:"keyword"`
	SearchedAt time.Time `gorm:"index;not null" json:"searchedAt"`
}

func (SearchHistory) TableName() string { return "gb_search_histories" }

type SearchEvent struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index" json:"-"`
	EventType  string    `gorm:"size:32;not null" json:"eventType"`
	Query      string    `gorm:"size:128;not null" json:"query"`
	ResultType string    `gorm:"size:16" json:"resultType"`
	ResourceID string    `gorm:"size:128" json:"resourceId"`
	Position   int       `json:"position"`
	OccurredAt time.Time `gorm:"index" json:"occurredAt"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (SearchEvent) TableName() string { return "gb_search_events" }
