// Package database
package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"test-psql/pkg/env"
)

func New(e *env.Env, gormConfig *gorm.Config) (*gorm.DB, error) {
	if e == nil {
		return nil, fmt.Errorf("env config is nil")
	}
	if gormConfig == nil {
		gormConfig = &gorm.Config{}
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		e.DBHost, e.DBPort, e.DBUser, e.DBPassword, e.DBName, e.DBSSLMode, e.DBTimezone)

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("db open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	sqlDB.SetMaxOpenConns(e.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(e.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(e.DBConnMaxLifetime)

	return db, nil
}
