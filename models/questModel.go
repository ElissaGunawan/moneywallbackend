package models

import (
	"time"

	"gorm.io/gorm"
)

type Quest struct {
	gorm.Model
	UserID      int
	QuestCode   string
	QuestName   string
	Exp         int
	CompletedAt *time.Time
	Reward      string
}
