package models

import (
	"time"

	"gorm.io/gorm"
)

type Expense struct {
	gorm.Model
	UserID      int
	Date        time.Time
	AccountID   int
	CategoryID  int
	Amount      int
	ExpenseName string
}

type ExpenseDashboard struct {
	CategoryID   int
	CategoryName string
	Amount       int
}
