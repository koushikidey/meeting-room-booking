package config

import (
	"log"

	"os"

	"github.com/joho/godotenv"
	"github.com/koushikidey/go-meetingroombook/pkg/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("Info: .env file not loaded, using system environment variables")
	}
	// err := godotenv.Load(".env")
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }
}

var (
	db *gorm.DB
)

func GetDB() *gorm.DB {
	return db
}

func MigrateDB(db *gorm.DB) {
	db.AutoMigrate(&models.Room{}, &models.Employee{}, &models.Booking{}, &models.GoogleToken{})
}
func Connect() {
	dsn := os.Getenv("DB_DSN")

	d, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	db = d
	MigrateDB(db)
}