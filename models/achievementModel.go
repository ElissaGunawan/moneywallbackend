package models

import (
	"time"

	"gorm.io/gorm"
)

type Achievement struct {
	gorm.Model
	UserID          int
	AchievementCode string
	AchievementName string
	Exp             int
	CompletedAt     *time.Time
	Reward          string
	ReduceExpense   float32
}
