package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"golang.org/x/oauth2"

	"github.com/mspcix/google-classroom-course-downloader/database"
	"github.com/mspcix/google-classroom-course-downloader/models"
	"github.com/mspcix/google-classroom-course-downloader/utils"
)

// Return courses that aren't present in the db.
func FilterNewCourses(classrooms []models.Course) ([]models.Course, error) {
	newClassrooms := []models.Course{}

	existingClassroomsIDs, err := database.GetCoursesGCIDs()
	if err != nil {
		return newClassrooms, err
	}

	// Convert the existingClassroomsID slice into a map for faster lookups.
	existingIDsMap := make(map[string]bool)
	for _, id := range existingClassroomsIDs {
		existingIDsMap[id] = true
	}

	for _, classroom := range classrooms {
		if !existingIDsMap[classroom.GCID] {
			newClassrooms = append(newClassrooms, classroom)
		}
	}

	return newClassrooms, nil
}

// Fetch the classrooms for the user using Google Classroom API
func GetCoursesFromAPI(r *http.Request, token string) ([]models.Course, error) {
	httpClient := utils.OAuthConfig.Client(r.Context(), &oauth2.Token{AccessToken: token})

	// Makes a GET request to the Classroom API to retrieve the list of classrooms
	response, err := httpClient.Get("https://classroom.googleapis.com/v1/courses")
	if err != nil {
		return nil, fmt.Errorf("error getting classrooms: %w", err)
	}
	defer response.Body.Close()

	// Parse the response body to get the list of classrooms
	var classroomsResponse struct {
		Courses []models.Course `json:"courses"`
	}
	err = json.NewDecoder(response.Body).Decode(&classroomsResponse)
	if err != nil {
		return nil, fmt.Errorf("error decoding classrooms response: %w", err)
	}

	// Populate the UserID field (foreign key) of each course
	gcuid, err := database.GetGCUIDByToken(token)
	if err != nil {
		return nil, err
	}

	for i := range classroomsResponse.Courses {
		classroomsResponse.Courses[i].UserGCID = gcuid
	}

	return classroomsResponse.Courses, nil
}

// Fetch all announcements of a list of courses using Google Classroom API
func GetAnnouncements(r *http.Request, token string, coursesIDs []string) ([]models.Announcement, error) {
	httpClient := utils.OAuthConfig.Client(r.Context(), &oauth2.Token{AccessToken: token})

	announcements := []models.Announcement{}
	for _, courseID := range coursesIDs {
		nextPageToken := ""
		for {
			// Make a GET request to the Classroom API to retrieve the list of announcements
			url := fmt.Sprintf("https://classroom.googleapis.com/v1/courses/%s/announcements", courseID)
			if nextPageToken != "" {
				url += "?pageToken=" + nextPageToken
			}

			response, err := httpClient.Get(url)
			if err != nil {
				log.Printf("error getting announcements: %v", err)
				return nil, err
			}
			defer response.Body.Close()

			// Parse the response body to get the list of announcements
			var announcementsResponse struct {
				Announcements []models.Announcement `json:"announcements"`
				NextPageToken string                `json:"nextPageToken"`
			}
			err = json.NewDecoder(response.Body).Decode(&announcementsResponse)
			if err != nil {
				log.Printf("error decoding announcements response: %v", err)
				return nil, err
			}

			// Set the title, type and url of materials
			for i := range announcementsResponse.Announcements {
				for j := range announcementsResponse.Announcements[i].Materials {
					announcementsResponse.Announcements[i].Materials[j].SetTitleTypeURL()
				}
			}

			announcements = append(announcements, announcementsResponse.Announcements...)

			// Check if there's a nextPageToken and continue fetching next page
			nextPageToken = announcementsResponse.NextPageToken
			if nextPageToken == "" {
				break
			}
		}
	}

	return announcements, nil
}

// Fetch the coursework materials of a list of courses using Google Classroom API
func GetCourseWorkMaterials(r *http.Request, token string, courseIDs []string) ([]models.CourseWorkMaterial, error) {
	httpClient := utils.OAuthConfig.Client(r.Context(), &oauth2.Token{AccessToken: token})

	var allCourseWorkMaterials []models.CourseWorkMaterial

	for _, courseID := range courseIDs {
		nextPageToken := ""
		for {
			// Make a GET request to the Classroom API to retrieve the list of coursework materials
			url := fmt.Sprintf("https://classroom.googleapis.com/v1/courses/%s/courseWorkMaterials?pageSize=50&pageToken=%s", courseID, nextPageToken)
			response, err := httpClient.Get(url)
			if err != nil {
				return nil, err
			}
			defer response.Body.Close()

			// Parse the response body to get the list of coursework materials
			var courseWorkMaterialsResponse struct {
				CourseWorkMaterials []models.CourseWorkMaterial `json:"courseWorkMaterial"`
				NextPageToken       string                      `json:"nextPageToken"`
			}
			err = json.NewDecoder(response.Body).Decode(&courseWorkMaterialsResponse)
			if err != nil {
				return nil, err
			}

			// Set the title, type and url of materials
			for i := range courseWorkMaterialsResponse.CourseWorkMaterials {
				for j := range courseWorkMaterialsResponse.CourseWorkMaterials[i].Materials {
					courseWorkMaterialsResponse.CourseWorkMaterials[i].Materials[j].SetTitleTypeURL()
				}
			}

			allCourseWorkMaterials = append(allCourseWorkMaterials, courseWorkMaterialsResponse.CourseWorkMaterials...)

			// Check if there are more pages to fetch
			if courseWorkMaterialsResponse.NextPageToken == "" {
				break
			}
			nextPageToken = courseWorkMaterialsResponse.NextPageToken
		}
	}

	return allCourseWorkMaterials, nil
}

// Download courses' materials from links in the database
func DownloadCourses(coursesIDs []string, token *string) error {
	log.Printf("Downloading %v course(s)...", len(coursesIDs))
	courses, err := database.GetCoursesByIDs(coursesIDs)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var downloadItems []models.DownloadItem
	maxConcurrentDownloadsStr := os.Getenv("MAX_CONCURRENT_DOWNLOADS")
	maxConcurrentDownloads, err := strconv.Atoi(maxConcurrentDownloadsStr)
	if err != nil {
		return err
	}
	semaphore := make(chan struct{}, maxConcurrentDownloads)

	for _, course := range courses {
		courseDownloadItems := course.GetDownloadItems()
		downloadItems = append(downloadItems, courseDownloadItems...)
	}

	// Create a channel to signal when the download is complete
	downloadCompleteCh := make(chan struct{})

	// Start the token refreshing goroutine
	go func() {
		defer close(downloadCompleteCh)
		ticker := time.NewTicker(20 * time.Minute) // Refresh token every 20 minutes
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				RefreshToken(token)
			case <-downloadCompleteCh:
				return
			}
		}
	}()

	for _, item := range downloadItems {
		item := item // Capture range variable
		wg.Add(1)
		go func(item models.DownloadItem) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := os.MkdirAll(item.DownloadFolderPath, os.ModePerm); err != nil {
				log.Printf("error creating folder: %v", err)
			}

			// Save materials and download files
			if err := saveDownloadItem(item, token); err != nil {
				log.Printf("error saving materials: %v", err)
			}
		}(item)
	}

	// Wait for downloads to complete
	wg.Wait()

	// Signal that the download is complete, stopping the token refreshing goroutine
	downloadCompleteCh <- struct{}{}

	log.Println("Finished downloading courses")
	return nil
}

func saveDownloadItem(item models.DownloadItem, token *string) error {
	if item.Text != "" {
		err := saveItemText(item.DownloadFolderPath, item.Text)
		if err != nil {
			log.Printf("error saving text: %v", err)
		}
	}

	for _, material := range item.Materials {
		switch material.Type {
		case "youtubeVideo", "link":
			if err := saveLinkToFile(item.DownloadFolderPath, material.URL); err != nil {
				log.Printf("error saving link: %v", err)
			}
		case "driveFile":
			if err := saveDriveFile(item.DownloadFolderPath, token, material); err != nil {
				log.Printf("error saving drive file: %v", err)
			}
		}
	}
	return nil
}

func saveItemText(FolderPath, text string) error {
	filePath := filepath.Join(FolderPath, "Announcement.txt")
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(text)
	return err
}

func saveLinkToFile(folderPath, link string) error {
	filePath := filepath.Join(folderPath, "links.txt")
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(link + "\n")
	return err
}

func saveDriveFile(folderPath string, token *string, material models.Material) error {
	filePath := filepath.Join(folderPath, utils.RemoveInvalidChars(material.Title))
	fileID, err := database.GetDriveFileID(material.ID)
	if err != nil {
		log.Printf("error retrieving fileID: %v", err)
		return err
	}

	if fileID == "" {
		fileID, err = database.GetDriveFileIDByTitle(material.Title)
		if err != nil {
			log.Printf("error retrieving fileID from material Title: %v", err)
			return err
		}
	}

	err = utils.DownloadDriveFile(token, fileID, filePath)
	if err != nil {
		log.Printf("error downloading material: %v", err)
	}

	return nil
}
