// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package main

import (
	"geregetemplateai/cmd/seed/seeders"
	"geregetemplateai/internal/config"
	"geregetemplateai/internal/constants"
	"geregetemplateai/internal/datasources/drivers"
	"geregetemplateai/pkg/logger"
)

func init() {
	if err := config.InitializeAppConfig(); err != nil {
		logger.Fatal(err.Error(), logger.Fields{constants.LoggerCategory: constants.LoggerCategoryConfig})
	}
	logger.Info("configuration loaded", logger.Fields{constants.LoggerCategory: constants.LoggerCategoryConfig})
}

func main() {
	// Production-д seed-ийг хэзээ ч ажиллуулахгүй — таамаглахад хялбар нууц
	// үгтэй admin backdoor суулгахаас сэргийлнэ.
	if config.AppConfig.Environment == constants.EnvironmentProduction {
		logger.Fatal("seed is disabled in production", logger.Fields{constants.LoggerCategory: constants.LoggerCategorySeeder})
	}

	db, err := drivers.SetupGORMPostgres()
	if err != nil {
		logger.Panic(err.Error(), logger.Fields{constants.LoggerCategory: constants.LoggerCategorySeeder})
	}
	defer func() {
		if sqlDB, dbErr := db.DB(); dbErr == nil {
			_ = sqlDB.Close()
		}
	}()

	logger.Info("seeding...", logger.Fields{constants.LoggerCategory: constants.LoggerCategorySeeder})

	seeder := seeders.NewSeeder(db)
	err = seeder.UserSeeder(seeders.UserData)
	if err != nil {
		logger.Panic(err.Error(), logger.Fields{constants.LoggerCategory: constants.LoggerCategorySeeder})
	}

	logger.Info("seeding success!", logger.Fields{constants.LoggerCategory: constants.LoggerCategorySeeder})
}
