package postgres

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const defaultValue = "messagio"

var instance *gorm.DB

func New() *gorm.DB {
	if instance == nil {
		user := os.Getenv("POSTGRES_USER")
		dbName := os.Getenv("POSTGRES_DB")
		password := os.Getenv("POSTGRES_PASSWORD")
		host := os.Getenv("POSTGRES_HOST")
		port := os.Getenv("POSTGRES_PORT")

		if user == "" {
			user = defaultValue
		}

		if dbName == "" {
			dbName = defaultValue
		}

		if password == "" {
			password = defaultValue
		}

		if host == "" {
			host = "localhost"
		}

		if port == "" {
			port = "5432"
		}

		db, err := gorm.Open(postgres.New(postgres.Config{
			DSN: fmt.Sprintf(
				"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable TimeZone=UTC",
				user, password, dbName, host, port,
			),
			PreferSimpleProtocol: true,
		}), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})

		if err != nil {
			log.Fatal(err)
		}

		instance = db
	}

	return instance
}
