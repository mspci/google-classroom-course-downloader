package database

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mspcix/google-classroom-course-downloader/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func GetDB() *gorm.DB {
	return db
}

func InitDB() (*gorm.DB, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	// Local database connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	// Remote database connection string
	// connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=%s",
	// 	os.Getenv("EDB_HOST"), os.Getenv("EDB_PORT"), os.Getenv("EDB_USER"), os.Getenv("EDB_PASSWORD"), os.Getenv("EDB_NAME"), os.Getenv("TIME_ZONE"))

	// connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s sslmode=verify-full",
	// 	os.Getenv("NDB_USER"), os.Getenv("NDB_PASSWORD"), os.Getenv("NDB_NAME"), os.Getenv("NDB_HOST"))

	// Open a connection to the database using GORM
	var err error
	db, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  connStr,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
		// Logger:                 utils.DBLogger.LoggerInterface,
	})
	if err != nil {
		return nil, fmt.Errorf("error connecting to PostgreSQL: %w", err)
	}

	// Disable Logger to suppress GORM logging output for this operation
	// db.Logger = logger.Default.LogMode(logger.Silent)

	if err := db.AutoMigrate(&models.User{}, &models.Course{}, &models.Announcement{}, &models.Material{}, &models.DriveFile{}, &models.YoutubeVideo{}, &models.Link{}, &models.Form{}, &models.CourseWorkMaterial{}); err != nil {
		return nil, fmt.Errorf("error automigrating models: %w", err)
	}

	return db, nil
}
