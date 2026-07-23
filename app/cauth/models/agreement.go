package models

import "time"

type Agreement struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Version     string    `gorm:"size:32;uniqueIndex;not null" json:"version"`
	Title       string    `gorm:"size:128;not null" json:"title"`
	ContentURL  string    `gorm:"size:512" json:"contentUrl"`
	Summary     string    `gorm:"size:512" json:"summary"`
	Status      string    `gorm:"size:16;not null;default:published;index" json:"status"`
	PublishedAt time.Time `json:"publishedAt"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (Agreement) TableName() string { return "gb_agreements" }

type PasswordReset struct {
	ID             uint       `gorm:"primaryKey"`
	AccountHash    string     `gorm:"size:64;index;not null"`
	CodeHash       string     `gorm:"size:64;not null"`
	ExpiresAt      time.Time  `gorm:"not null"`
	UsedAt         *time.Time
	FailedAttempts int        `gorm:"not null;default:0"`
	CreatedAt      time.Time
}

func (PasswordReset) TableName() string { return "gb_password_resets" }

type DeviceAccount struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	DeviceID      string    `gorm:"size:128;index;not null" json:"deviceId"`
	UserID        uint      `gorm:"index;not null" json:"userId"`
	AccountMasked string    `gorm:"size:64;not null" json:"accountMasked"`
	LastLoginAt   time.Time `json:"lastLoginAt"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (DeviceAccount) TableName() string { return "gb_device_accounts" }
