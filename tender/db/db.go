package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"tender/models"
)

var DB *gorm.DB

func Connect() {
	var err error

	dsn := fmt.Sprintf("host=%s user=%s password=%s database=%s port=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		"5432")

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to Postgres", err)
	}

	err = DB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error
	if err != nil {
		log.Fatalf("Failed to enable uuid-ossp extension: %v", err)
	}

	err = DB.Exec(`CREATE TYPE organization_type AS ENUM ('IE', 'LLC', 'JSC')`).Error
	if err != nil {
		log.Fatalf("failed to create enum type: %v", err)
	}

	err = DB.Exec(`CREATE TYPE tender_status AS ENUM ('CREATED', 'PUBLISHED', 'CLOSED')`).Error
	if err != nil {
		log.Fatalf("failed to create enum type: %v", err)
	}
	err = DB.Exec(`CREATE TYPE tender_service_type AS ENUM ('CONSTRUCTION', 'DELIVERY', 'MANUFACTURE')`).Error
	if err != nil {
		log.Fatalf("failed to create enum type: %v", err)
	}

	err = DB.Exec(`CREATE TYPE bid_status AS ENUM ('CREATED', 'PUBLISHED', 'CANCELED')`).Error
	if err != nil {
		log.Fatalf("failed to create enum type: %v", err)
	}

	err = DB.Exec(`CREATE TYPE bid_author_type AS ENUM ('ORGANIZATION', 'USER')`).Error
	if err != nil {
		log.Fatalf("failed to create enum type: %v", err)
	}
}

func Migrate() {
	DB.AutoMigrate(&models.Employee{}, &models.Organization{}, &models.OrganizationResponsible{}, &models.Tender{}, &models.TenderVersion{}, &models.Bid{}, &models.BidVersion{})
}
