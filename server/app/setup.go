// Package app preforms basic app functions like setup,Build time variables, loading config registering controllers,starting http server and global definitions
package app

import (
	"fmt"
	"treblle/model"
	"time"

	"go.uber.org/dig"
	"go.uber.org/zap"
)

// Setup will preform app setup or panic of it fails
// Can only be called once
func Setup() {
	// Logger setup
	{
		var err error

		if Build == BuildDev {
			err = devLoggerSetup()
			if err != nil {
				fmt.Printf("err: %v\n", err)
				panic("failed to setup logger")
			}
		} else {
			err = prodLoggerSetup()
			if err != nil {
				fmt.Printf("err: %v\n", err)
				panic("failed to setup logger")
			}
		}
	}

	// Print build time variables
	{
		zap.S().Infof("Build:      \t\t %s", Build)
		zap.S().Infof("Version:    \t\t %s", Version)
		zap.S().Infof("Commit Hash:\t\t %s", CommitHash)
		zap.S().Infof("Build Time Stamp:\t %s", BuildTimestamp)
		zap.S().Sync()
	}

	LoadConfig()

	// Dig setup
	{
		digContainer = dig.New()
	}

	// gorm and Database setup
	{
		db := newDbConn()
		sqlDB, err := db.DB()
		if err != nil {
			zap.S().Panicf("failed to get database connection: %+v", err)
		}

		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)

		if err = db.AutoMigrate(model.GetAllModels()...); err != nil {
			zap.S().Panicf("Can't run AutoMigrate err = %+v", err)
		}

		Provide(newDbConn)
	}
}
