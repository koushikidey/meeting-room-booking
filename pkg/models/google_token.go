package models

import (
	"time"

	"gorm.io/gorm"
)

type GoogleToken struct {
	gorm.Model
	ID           uint      `gorm:"primaryKey"`
	EmployeeID   uint      `gorm:"not null"`
	AccessToken  string    `gorm:"type:text;not null"`
	RefreshToken string    `gorm:"type:text;not null"`
	Expiry       time.Time `gorm:"not null"`
}