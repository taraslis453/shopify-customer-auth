package postgresql

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgreSQLGorm struct {
	DB *gorm.DB
}

type Config struct {
	User     string
	Password string
	Host     string
	Database string
}

func NewPostgreSQLGorm(cfg Config) (*PostgreSQLGorm, error) {
	// connect to database
	dsn := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s",
		cfg.User, cfg.Password, cfg.Database, cfg.Host,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgresql: %s", err)
	}

	// create UUID extension.
	err = db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create uuid-ossp extension: %s", err)
	}

	return &PostgreSQLGorm{DB: db}, nil
}
