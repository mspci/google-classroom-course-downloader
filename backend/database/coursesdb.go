package database

import (
	"fmt"
	"log"

	"github.com/mspcix/google-classroom-downloader/models"
	"gorm.io/gorm"
)

// Insert a list of courses into the database.
func InsertCourses(courses []models.Course) error {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v\n", r)
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		log.Printf("Error starting transaction: %v\n", err)
		return err
	}

	batchSize := 50
	result := tx.CreateInBatches(courses, batchSize)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	return tx.Commit().Error
}

// Retrieve a paginated list of courses' IDs from the database.
func GetCoursesIDs() ([]string, error) {
	var coursesIDs []string
	result := db.Model(&models.Course{}).Pluck("id", &coursesIDs)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			fmt.Println("No courses found")
			return nil, nil // No courses found
		}
		return nil, fmt.Errorf("error retrieving courses IDs from the database: %w", result.Error)
	}
	return coursesIDs, nil
}

// Retrieve the URLs of the materials of a specific user
func GetMaterialsURLsByUser(token string) ([]string, error) {
	var materialsURLs []string

	// Fetch the user by their token
	var user models.User
	if err := db.Where("token = ?", token).Preload("Courses.Announcements.Materials").First(&user).Error; err != nil {
		return nil, err
	}

	// Extract materials' URLs
	for _, course := range user.Courses {
		for _, announcement := range course.Announcements {
			for _, material := range announcement.Materials {
				materialsURLs = append(materialsURLs, material.URL)
			}
		}
	}

	return materialsURLs, nil
}

// Returns all Courses of a user
func GetCoursesByToken(token string) ([]models.Course, error) {
	// Fetch the user by their token
	var user models.User
	if err := db.Where("token = ?", token).Preload("Courses.Announcements.Materials").First(&user).Error; err != nil {
		return nil, err
	}

	return user.Courses, nil
}

// Retrieves courses with the given course IDs
func GetCoursesByIDs(coursesIDs []string) ([]models.Course, error) {
	var courses []models.Course
	if err := db.Where("id IN ?", coursesIDs).Preload("Announcements.Materials").Find(&courses).Error; err != nil {
		return nil, err
	}

	return courses, nil
}

// Retrieves the drive file ID from a material's ID
func GetDriveFileID(materialID uint) (string, error) {
	var driveFileID string
	result := db.Model(&models.DriveFile{}).Where("material_id = ?", materialID).Pluck("drive_file_drive_file_id", &driveFileID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			fmt.Println("No drive file found")
			return "", nil // No drive file found
		}
		return "", fmt.Errorf("error retrieving drive file ID from the database: %w", result.Error)
	}
	return driveFileID, nil
}
