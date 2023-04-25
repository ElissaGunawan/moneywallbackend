package models

import (
	"time"

	"gorm.io/gorm"
)

type Income struct {
	gorm.Model
	UserID     int
	Date       time.Time
	AccountID  int
	IncomeName string
	Amount     int
}

type IncomeDashboard struct {
	AccountID   int
	AccountName string
	Amount      int
}
