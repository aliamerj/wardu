package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/aliamerj/wardu/shared/env"
	"github.com/aliamerj/wardu/shared/models"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

type service struct {
	db *gorm.DB
}

var (
	database   = env.GetString("POSTGRES_DATABASE", "wardu")
	password   = env.GetString("POSTGRES_PASSWORD", "wardu")
	username   = env.GetString("POSTGRES_USERNAME", "wardu")
	port       = env.GetString("POSTGRES_PORT", "5432")
	host       = env.GetString("POSTGRES_HOST", "postgres.default.svc.cluster.local")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s", host, username, password, database, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`SELECT pg_advisory_xact_lock(hashtext('wardu:database:migrate'))`).Error; err != nil {
			return err
		}

		return tx.AutoMigrate(&models.Namespace{})
	}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	dbInstance = &service{
		db: db,
	}
	return dbInstance
}
