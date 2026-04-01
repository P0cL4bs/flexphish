package storage

import (
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewDatabase(path string) (*gorm.DB, error) {
	dsn := path
	if strings.Contains(path, "?") {
		dsn = path + "&_foreign_keys=on"
	} else {
		dsn = path + "?_foreign_keys=on"
	}
	return gorm.Open(sqlite.Open(dsn), &gorm.Config{})
}
