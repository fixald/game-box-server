package models

import "time"

type User struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	PhoneHash         string     `gorm:"size:64;uniqueIndex;not null" json:"-"`
	PhoneCiphertext   string     `gorm:"size:255;not null" json:"-"`
	Email             string     `gorm:"size:255" json:"-"`
	Account           string     `gorm:"size:128;index" json:"-"`
	AccountHash       string     `gorm:"size:64;index" json:"-"`
	PasswordHash      string     `gorm:"size:128" json:"-"`
	PasswordUpdatedAt *time.Time `json:"-"`
	Nickname          string     `gorm:"size:64" json:"nickname"`
	AvatarURL         string     `gorm:"size:512" json:"avatarUrl"`
	Status            string     `gorm:"size:16;not null;default:active;index" json:"status"`
	VipLevel          int        `gorm:"not null;default:0" json:"vipLevel"`
	Points            int        `gorm:"not null;default:0" json:"points"`
	TokenVersion      int64      `gorm:"not null;default:1" json:"-"`
	AgreementVersion  string     `gorm:"size:32" json:"agreementVersion"`
	RealNameStatus    string     `gorm:"size:16;not null;default:unverified" json:"realNameStatus"`
	LastLoginAt       *time.Time `json:"lastLoginAt"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type VIPLevel struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"size:16;uniqueIndex;not null" json:"name"`
	Requirement string `gorm:"size:128;not null" json:"requirement"`
	Growth      int    `gorm:"not null;default:0" json:"growth"`
	Description string `gorm:"size:255" json:"desc"`
	Status      string `gorm:"size:16;not null;default:active" json:"-"`
	Sort        int    `gorm:"not null;default:0" json:"-"`
}

func (VIPLevel) TableName() string { return "gb_vip_levels" }

func (User) TableName() string { return "gb_users" }

type SMSCode struct {
	ID             uint   `gorm:"primaryKey"`
	PhoneHash      string `gorm:"size:64;index;not null"`
	CodeHash       string `gorm:"size:64;not null"`
	ExpiresAt      time.Time
	UsedAt         *time.Time
	FailedAttempts int
	LockedUntil    *time.Time
	CreatedAt      time.Time
}

func (SMSCode) TableName() string { return "gb_sms_codes" }

type RefreshToken struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"index;not null" json:"userId"`
	TokenHash  string     `gorm:"size:64;uniqueIndex;not null" json:"-"`
	ExpiresAt  time.Time  `gorm:"index;not null" json:"expiresAt"`
	RevokedAt  *time.Time `json:"revokedAt"`
	ReplacedBy string     `gorm:"size:64" json:"-"`
	CreatedAt  time.Time  `json:"createdAt"`
}

func (RefreshToken) TableName() string { return "gb_refresh_tokens" }

type FavoriteGame struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"uniqueIndex:ux_favorite_game"`
	GameID    uint `gorm:"uniqueIndex:ux_favorite_game"`
	CreatedAt time.Time
}

func (FavoriteGame) TableName() string { return "gb_favorite_games" }

type RecentGame struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"index"`
	GameID    uint
	ServerID  uint
	VisitedAt time.Time `gorm:"index"`
}

func (RecentGame) TableName() string { return "gb_recent_games" }

type RecentServer struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"index"`
	ServerID  uint
	VisitedAt time.Time `gorm:"index"`
}

func (RecentServer) TableName() string { return "gb_recent_servers" }

type DownloadRecord struct {
	ID           uint `gorm:"primaryKey"`
	UserID       uint `gorm:"index"`
	GameID       uint
	FileName     string    `gorm:"size:255"`
	Version      string    `gorm:"size:64"`
	Size         string    `gorm:"size:32"`
	Status       string    `gorm:"size:24;not null;default:completed"`
	Progress     int       `gorm:"not null;default:100"`
	DownloadedAt time.Time `gorm:"index"`
}

type CheckinRecord struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"uniqueIndex:ux_checkin_user_day;not null"`
	CheckinAt time.Time `gorm:"uniqueIndex:ux_checkin_user_day;not null"`
}

func (CheckinRecord) TableName() string { return "gb_checkin_records" }

func (DownloadRecord) TableName() string { return "gb_download_records" }

type Message struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"index"`
	Type      string `gorm:"size:24;not null;default:system"`
	Title     string `gorm:"size:128"`
	Content   string `gorm:"type:text"`
	ReadAt    *time.Time
	CreatedAt time.Time `gorm:"index"`
}

func (Message) TableName() string { return "gb_messages" }

type LoginRecord struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	IP        string    `gorm:"size:64"`
	UserAgent string    `gorm:"size:512"`
	LoginAt   time.Time `gorm:"index"`
	Success   bool
}

func (LoginRecord) TableName() string { return "gb_login_records" }

type BindCode struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"index"`
	Kind      string `gorm:"size:16"`
	Target    string `gorm:"size:255"`
	CodeHash  string `gorm:"size:64"`
	ExpiresAt time.Time
	UsedAt    *time.Time
}

func (BindCode) TableName() string { return "gb_bind_codes" }

type UserSettings struct {
	ID               uint `gorm:"primaryKey"`
	UserID           uint `gorm:"uniqueIndex;not null"`
	ShowOnlineStatus bool `gorm:"not null;default:true"`
	AllowMessages    bool `gorm:"not null;default:true"`
	NotifySystem     bool `gorm:"not null;default:true"`
	NotifyActivity   bool `gorm:"not null;default:true"`
	NotifyLive       bool `gorm:"not null;default:true"`
	UpdatedAt        time.Time
}

func (UserSettings) TableName() string { return "gb_user_settings" }

type RewardRecord struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index;not null"`
	Name      string    `gorm:"size:128"`
	Code      string    `gorm:"size:64"`
	Status    string    `gorm:"size:24;not null;default:claimed"`
	ClaimedAt time.Time `gorm:"index"`
}

type TaskClaim struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"uniqueIndex:ux_task_claim_user_task;not null"`
	TaskID    string    `gorm:"size:64;uniqueIndex:ux_task_claim_user_task;not null"`
	Points    int       `gorm:"not null;default:0"`
	ClaimedAt time.Time `gorm:"index"`
}

func (TaskClaim) TableName() string { return "gb_task_claims" }

type Task struct {
	ID          uint      `gorm:"primaryKey" json:"-"`
	Code        string    `gorm:"size:64;uniqueIndex;not null" json:"id"`
	Category    string    `gorm:"size:24;index;not null" json:"category"`
	Title       string    `gorm:"size:128;not null" json:"title"`
	Description string    `gorm:"size:255" json:"description"`
	Icon        string    `gorm:"size:16" json:"icon"`
	Target      int       `gorm:"not null;default:1" json:"target"`
	Points      int       `gorm:"not null;default:0" json:"-"`
	ActionLabel string    `gorm:"size:32" json:"actionLabel"`
	ActionRoute string    `gorm:"size:128" json:"actionRoute"`
	Status      string    `gorm:"size:16;not null;default:active;index" json:"-"`
	Sort        int       `gorm:"not null;default:0" json:"-"`
	Rewards     string    `gorm:"type:text" json:"rewards"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (Task) TableName() string { return "gb_tasks" }

type CheckinReward struct {
	ID        uint   `gorm:"primaryKey"`
	Level     int    `gorm:"uniqueIndex;not null" json:"level"`
	Name      string `gorm:"size:64;not null" json:"name"`
	Reward    string `gorm:"size:128;not null" json:"reward"`
	Icon      string `gorm:"size:16" json:"icon"`
	Status    string `gorm:"size:16;not null;default:active" json:"-"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (CheckinReward) TableName() string { return "gb_checkin_rewards" }

type CheckinRewardClaim struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"uniqueIndex:ux_checkin_reward_claim_user_level;not null"`
	Level     int       `gorm:"uniqueIndex:ux_checkin_reward_claim_user_level;not null"`
	ClaimedAt time.Time `gorm:"index"`
}

func (CheckinRewardClaim) TableName() string { return "gb_checkin_reward_claims" }

func (RewardRecord) TableName() string { return "gb_reward_records" }
