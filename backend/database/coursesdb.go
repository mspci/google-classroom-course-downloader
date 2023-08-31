package database

import (
	"fmt"

	"github.com/mspcix/google-classroom-course-downloader/models"
	"gorm.io/gorm"
)

// Insert a list of courses into the database.
func SaveCourses(courses []models.Course) (err error) {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			err = fmt.Errorf("panic recovered: %v", r)
		}
	}()

	if err = tx.Error; err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	batchSize := 50
	result := tx.CreateInBatches(courses, batchSize)
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("error creating courses: %w", result.Error)
	}

	if err = tx.Commit().Error; err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// Retrieve a paginated list of courses' IDs from the database.
func GetCoursesGCIDs() ([]string, error) {
	var coursesIDs []string
	result := db.Model(&models.Course{}).Pluck("gcid", &coursesIDs)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			fmt.Println("No courses found")
			return nil, nil // No courses found
		}
		return nil, fmt.Errorf("error retrieving courses IDs from the database: %w", result.Error)
	}
	return coursesIDs, nil
}

// Returns all Courses of a user
func GetCoursesByGCUID(gcuid string) ([]models.Course, error) {
	// Fetch the user by their token
	var courses []models.Course
	if err := db.Where("user_gcid_f = ?", gcuid).Preload("Announcements.Materials").Preload("CourseWorkMaterials.Materials").Find(&courses).Error; err != nil {
		return nil, err
	}

	return courses, nil
}

// Retrieves courses with the given course IDs
func GetCoursesByIDs(coursesIDs []string) ([]models.Course, error) {
	var courses []models.Course

	if err := db.Where("gcid IN ?", coursesIDs).Preload("Announcements.Materials").Preload("CourseWorkMaterials.Materials").Find(&courses).Error; err != nil {
		return nil, err
	}

	return courses, nil
}

// Retrieves the drive file ID from a material's ID
func GetDriveFileID(materialID uint) (string, error) {
	var driveFileID string
	result := db.Model(&models.DriveFile{}).Where("material_id_f = ?", materialID).Pluck("drive_file_drive_file_id", &driveFileID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			fmt.Println("No drive file found")
			return "", nil // No drive file found
		}
		return "", fmt.Errorf("error retrieving drive file ID from the database: %w", result.Error)
	}
	return driveFileID, nil
}

// Retrieves the drive file ID from a material's Title
func GetDriveFileIDByTitle(title string) (string, error) {
	var driveFileID string
	result := db.Model(&models.DriveFile{}).Where("drive_file_drive_file_title = ?", title).Pluck("drive_file_drive_file_id", &driveFileID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			fmt.Println("No drive file found")
			return "", nil // No drive file found
		}
		return "", fmt.Errorf("error retrieving drive file ID from the database: %w", result.Error)
	}
	return driveFileID, nil
}

// Retrieves the course name from a announcement's ID through announcement.course_id_f
func GetCourseNameByAnnouncementID(announcementID string) (string, error) {
	// Retrieve the associated course ID from the announcement
	var announcement models.Announcement
	if err := db.Model(&models.Announcement{}).Where("id = ?", announcementID).First(&announcement).Error; err != nil {
		return "", err
	}

	// Retrieve the course based on the course ID from the announcement
	var course models.Course
	if err := db.Model(&models.Course{}).Where("id = ?", announcement.CourseID).First(&course).Error; err != nil {
		return "", err
	}

	return course.Name, nil
}

func GetCourseNameByCourseWokrMaterialID(courseWorkMaterialID string) (string, error) {
	// Retrieve the associated course ID from the announcement
	var courseWorkMaterial models.CourseWorkMaterial
	if err := db.Model(&models.CourseWorkMaterial{}).Where("id = ?", courseWorkMaterialID).First(&courseWorkMaterial).Error; err != nil {
		return "", err
	}

	// Retrieve the course based on the course ID from the announcement
	var course models.Course
	if err := db.Model(&models.Course{}).Where("id = ?", courseWorkMaterial.CourseID).First(&course).Error; err != nil {
		return "", err
	}

	return course.Name, nil
}
