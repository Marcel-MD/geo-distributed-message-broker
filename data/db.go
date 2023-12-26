package data

import (
	"geo-distributed-message-broker/config"
	"log/slog"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewDB(cfg config.Config) (*gorm.DB, error) {
	slog.Info("Creating new database connection")

	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func CloseDB(db *gorm.DB) error {
	dbSql, err := db.DB()
	if err != nil {
		return err
	}

	return dbSql.Close()
}
