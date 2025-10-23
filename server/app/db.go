package app

import (
	gormzap "treblle/util/gormZap"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newDbConn() *gorm.DB {
	db, err := gorm.Open(postgres.Open(DbConn), &gorm.Config{
		// NOTE: change LogMode if needed when debugging
		Logger: gormzap.NewGormZapLogger().LogMode(logger.Warn),
	})
	if err != nil {
		zap.S().Panicf("failed to connect database err = %+v", err)
	}
	return db
}

func testDbConn() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file:db?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		zap.S().Panicf("failed to connect database err = %+v", err)
	}
	return db
}
