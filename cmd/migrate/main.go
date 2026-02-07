package main

import (
	"fmt"
	"os"

	"test-psql/internal/database"
	"test-psql/internal/migrations"
	"test-psql/pkg/env"
	"test-psql/pkg/logger"
)

func main() {
	if err := logger.Init(false, "migrate.log"); err != nil {
		fmt.Fprintf(os.Stderr, "logger init: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = logger.Close() }()

	cfg, err := env.LoadFromFile("config.env")
	if err != nil {
		logger.Error(fmt.Sprintf("config: %v", err))
		os.Exit(1)
	}

	db, err := database.New(cfg, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("database: %v", err))
		os.Exit(1)
	}
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	migrationsPath := "migrations"
	if len(os.Args) > 1 {
		migrationsPath = os.Args[1]
	}

	logger.Info(fmt.Sprintf("running migrations from %s", migrationsPath))
	if err := migrations.Run(db, migrationsPath); err != nil {
		logger.Error(fmt.Sprintf("migration: %v", err))
		os.Exit(1)
	}

	logger.Info("migrations completed successfully")
}
