package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	UserID       int
	CategoryName string
}
