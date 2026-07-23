package models

import "time"

type InviteCode struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Code      string     `gorm:"size:32;uniqueIndex;not null" json:"code"`
	Status    string     `gorm:"size:16;not null;default:active;index" json:"status"`
	MaxUses   int        `gorm:"not null;default:0" json:"maxUses"` // 0 = unlimited
	UsedCount int        `gorm:"not null;default:0" json:"usedCount"`
	ExpiresAt *time.Time `json:"expiresAt"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (InviteCode) TableName() string { return "gb_invite_codes" }

type RegisterTicket struct {
	ID        uint       `gorm:"primaryKey"`
	DeviceID  string     `gorm:"size:128;index;not null"`
	TokenHash string     `gorm:"size:64;uniqueIndex;not null"`
	ExpiresAt time.Time  `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt time.Time
}

func (RegisterTicket) TableName() string { return "gb_register_tickets" }

type RegisterEvent struct {
	ID        uint      `gorm:"primaryKey"`
	DeviceID  string    `gorm:"size:128;index;not null"`
	Account   string    `gorm:"size:128;not null"`
	ClientIP  string    `gorm:"size:64;index"`
	CreatedAt time.Time `gorm:"index"`
}

func (RegisterEvent) TableName() string { return "gb_register_events" }
