package models

import "time"

type UserBan struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"index;not null" json:"userId"`
	BanType    string     `gorm:"size:16;not null;index" json:"banType"`
	Reason     string     `gorm:"size:512;not null" json:"reason"`
	Source     string     `gorm:"size:32" json:"source"`
	StartsAt   time.Time  `json:"startsAt"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	Status     string     `gorm:"size:16;not null;default:active;index" json:"status"`
	OperatorID uint       `json:"operatorId"`
	LiftedAt   *time.Time `json:"liftedAt"`
	LiftedBy   uint       `json:"liftedBy"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

func (UserBan) TableName() string { return "gb_user_bans" }
