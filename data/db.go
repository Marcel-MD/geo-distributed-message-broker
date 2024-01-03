package data

import (
	"geo-distributed-message-broker/config"
	"log/slog"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB(cfg config.Config) (*gorm.DB, error) {
	slog.Info("Creating new database connection 💾")

	db, err := gorm.Open(sqlite.Open(cfg.Database), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&Message{})
	db.AutoMigrate(&MessageConsumedRecord{})

	return db, nil
}

func CloseDB(db *gorm.DB) error {
	dbSql, err := db.DB()
	if err != nil {
		return err
	}

	return dbSql.Close()
}
