package models

import "gorm.io/gorm"

type Account struct {
	gorm.Model
	UserID      int
	AccountName string
	Amount      int
	FirstAmount int
}
