package storage

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewDatabase(path string) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(path), &gorm.Config{})
}
