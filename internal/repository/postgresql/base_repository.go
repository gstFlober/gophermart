package postgresql

import (
	"gorm.io/gorm"
)

type BaseRepository struct {
	db *gorm.DB
}
