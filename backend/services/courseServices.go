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

	"golang.org/x/oauth2"

	"github.com/mspcix/google-classroom-downloader/database"
	"github.com/mspcix/google-classroom-downloader/models"
	"github.com/mspcix/google-classroom-downloader/utils"
)

// Return courses that aren't present in the db.
func FilterNewCourses(classrooms []models.Course) ([]models.Course, error) {
	newClassrooms := []models.Course{}

	existingClassroomsIDs, err := database.GetCoursesIDs()
	if err != nil {
		return newClassrooms, err
	}

	// Convert the existingClassroomsID slice into a map for faster lookups.
	existingIDsMap := make(map[string]bool)
	for _, id := range existingClassroomsIDs {
		existingIDsMap[id] = true
	}

	for _, classroom := range classrooms {
		if !existingIDsMap[classroom.ID] {
			newClassrooms = append(newClassrooms, classroom)
		}
	}

	return newClassrooms, nil
}

// Fetch the classrooms for the user using Google Classroom API
func GetCourses(r *http.Request, token string) ([]models.Course, error) {
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

			// Set the title and type of materials
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

// Download courses' materials from links in the database
func DownloadCourses(coursesIDs []string, token string) error {
	courses, err := database.GetCoursesByIDs(coursesIDs)
	if err != nil {
		return err
	}

	// Use a WaitGroup to wait for all goroutines to finish
	var downloadCompleteWG sync.WaitGroup

	// Limit concurrent downloads
	maxConcurrentDownloadsStr := os.Getenv("MAX_CONCURRENT_DOWNLOADS")
	maxConcurrentDownloads, err := strconv.Atoi(maxConcurrentDownloadsStr)
	if err != nil {
		return err
	}
	semaphore := make(chan struct{}, maxConcurrentDownloads)

	// Iterate over courses and announcements
	for _, course := range courses {
		courseFolderPath := filepath.Join(utils.DownloadPath, course.Name)
		if err := os.MkdirAll(courseFolderPath, os.ModePerm); err != nil {
			return err
		}

		for _, announcement := range course.Announcements {
			creationDate := utils.MakeFolderNameFromTime(announcement.CreationTime)
			announcementFolderPath := filepath.Join(courseFolderPath, creationDate)
			if err := os.MkdirAll(announcementFolderPath, os.ModePerm); err != nil {
				return err
			}

			announcementFilePath := filepath.Join(announcementFolderPath, "announcement.txt")
			if err := saveAnnouncementText(announcementFilePath, announcement.Text); err != nil {
				return err
			}

			downloadCompleteWG.Add(1)
			go func(materials []models.Material, announcementFolderPath, token string) {
				defer downloadCompleteWG.Done()

				for _, material := range materials {
					switch material.Type {
					case "youtubeVideo", "link":
						if err := saveLinkToFile(announcementFolderPath, material.URL); err != nil {
							log.Printf("error saving link: %v", err)
						}
					case "driveFile":
						semaphore <- struct{}{}
						downloadCompleteWG.Add(1)
						go func(material models.Material, announcementFolderPath, token string) {
							defer func() { <-semaphore }()
							defer downloadCompleteWG.Done()
							saveDriveFile(material, announcementFolderPath, token)
						}(material, announcementFolderPath, token)
					}
				}
			}(announcement.Materials, announcementFolderPath, token)
		}
	}

	downloadCompleteWG.Wait()
	return nil
}

func saveAnnouncementText(filePath, text string) error {
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

func saveDriveFile(material models.Material, folderPath string, token string) {
	filePath := filepath.Join(folderPath, material.Title)
	fileID, err := database.GetDriveFileID(material.ID)
	if err != nil {
		log.Printf("error retrieving fileID: %v", err)
		return
	}

	err = utils.DownloadDriveFile(token, fileID, filePath)
	if err != nil {
		log.Printf("error downloading material: %v", err)
	}
}
