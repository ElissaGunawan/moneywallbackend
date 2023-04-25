package models

import (
	"time"

	"gorm.io/gorm"
)

type Avatar struct {
	gorm.Model
	UserID     int
	AvatarCode string
	AvatarURL  string
	UnlockedAt *time.Time
}
