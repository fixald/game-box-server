package models

import "time"

type Report struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      *uint     `gorm:"index" json:"userId"`
	TargetType  string    `gorm:"size:32;not null" json:"targetType"`
	TargetID    string    `gorm:"size:64;not null" json:"targetId"`
	Reason      string    `gorm:"size:64;not null" json:"reason"`
	Detail      string    `gorm:"size:2000" json:"detail"`
	Status      string    `gorm:"size:20;not null;default:pending;index" json:"status"`
	HandlerID   uint      `json:"handlerId"`
	HandlerNote string    `gorm:"size:1000" json:"handlerNote"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (Report) TableName() string { return "reports" }

type Feedback struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        *uint     `gorm:"index" json:"userId"`
	Category      string    `gorm:"size:64;not null" json:"category"`
	Detail        string    `gorm:"size:4000;not null" json:"detail"`
	AttachmentIDs string    `gorm:"size:2000" json:"attachmentIds"`
	Status        string    `gorm:"size:20;not null;default:pending;index" json:"status"`
	HandlerID     uint      `json:"handlerId"`
	Result        string    `gorm:"size:2000" json:"result"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (Feedback) TableName() string { return "feedbacks" }
