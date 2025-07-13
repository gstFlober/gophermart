package entity

import (
	"time"
)

type OrderStatus string

const (
	OrderNew        OrderStatus = "NEW"
	OrderProcessing OrderStatus = "PROCESSING"
	OrderInvalid    OrderStatus = "INVALID"
	OrderProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	ID         uint        `gorm:"primaryKey;autoIncrement"`
	UserID     string      `gorm:"index;not null"`
	Number     string      `gorm:"uniqueIndex;not null"`
	Status     OrderStatus `gorm:"type:varchar(20);index;not null"`
	Accrual    float64     `gorm:"type:decimal(10,2);default:0.0"`
	UploadedAt time.Time   `gorm:"not null"`
	CreatedAt  time.Time   `gorm:"autoCreateTime"`
	UpdatedAt  time.Time   `gorm:"autoUpdateTime"`
}
