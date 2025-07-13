package entity

import (
	"time"
)

type Withdrawal struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	UserID      string    `gorm:"index;not null"`
	OrderNumber string    `gorm:"not null"`
	Sum         float64   `gorm:"type:decimal(10,2);not null"`
	ProcessedAt time.Time `gorm:"not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}
