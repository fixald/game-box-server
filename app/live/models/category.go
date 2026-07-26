package models

// Category is a client-visible live-room filter category.
type Category struct {
	ID      string `gorm:"primaryKey;size:64" json:"id"`
	Name    string `gorm:"size:64;not null" json:"name"`
	Type    string `gorm:"size:32;not null;index" json:"type"`
	Sort    int    `gorm:"not null;default:0;index" json:"sort"`
	Enabled bool   `gorm:"not null;default:true;index" json:"enabled"`
}

func (Category) TableName() string { return "gb_live_categories" }
