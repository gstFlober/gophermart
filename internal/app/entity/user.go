package entity

import (
	"time"
)

type User struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	Login          string    `gorm:"uniqueIndex;not null"`
	PasswordHash   string    `gorm:"not null"`
	CurrentBalance float64   `gorm:"type:decimal(10,2);default:0.0"`
	Withdrawn      float64   `gorm:"type:decimal(10,2);default:0.0"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}
