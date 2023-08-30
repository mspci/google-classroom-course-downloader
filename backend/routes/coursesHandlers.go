package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"github.com/mspcix/google-classroom-course-downloader/database"
	"github.com/mspcix/google-classroom-course-downloader/models"
	"github.com/mspcix/google-classroom-course-downloader/services"

	"github.com/mspcix/google-classroom-course-downloader/utils"
)

// Retrieves the list of new courses for the authenticated user from Google's Classroom API
// Inserts them into the database
func HandleDiscoverCourses(w http.ResponseWriter, r *http.Request, store sessions.Store) {
	startDiscovery := time.Now()
	log.Println("[HandleDiscoverCourses] /courses/discover hit")
	token, err := database.GetTokenFromSession(r, store)
	if err != nil {
		log.Println("Error retrieving token from the database:", err)
		return
	}

	courses, err := services.GetCoursesFromAPI(r, token)
	if err != nil {
		fmt.Println("Error retrieving courses:", err)
		http.Error(w, "Failed to retrieve courses", http.StatusInternalServerError)
		return
	}

	newCourses, err := services.FilterNewCourses(courses)
	if err != nil {
		fmt.Println("Error retrieving coursesId from the database:", err)
		http.Error(w, "Failed to filter new courses", http.StatusInternalServerError)
		return
	}

	newCoursesIDs := make([]string, len(newCourses))
	for i, course := range newCourses {
		newCoursesIDs[i] = course.GCID
	}

	announcements, err := services.GetAnnouncements(r, token, newCoursesIDs)
	if err != nil {
		fmt.Println("Error retrieving announcements:", err)
		http.Error(w, "Failed to retrieve announcements", http.StatusInternalServerError)
		return
	}

	courseWorkMaterials, err := services.GetCourseWorkMaterials(r, token, newCoursesIDs)
	if err != nil {
		fmt.Println("Error retrieving course work materials:", err)
		http.Error(w, "Failed to retrieve course work materials", http.StatusInternalServerError)
		return
	}

	// Create maps to store announcements and course work materials by course ID
	announcementsMap := make(map[string][]models.Announcement)
	for _, announcement := range announcements {
		announcementsMap[announcement.CourseID] = append(announcementsMap[announcement.CourseID], announcement)
	}
	courseWorkMaterialsMap := make(map[string][]models.CourseWorkMaterial)
	for _, courseWorkMaterial := range courseWorkMaterials {
		courseWorkMaterialsMap[courseWorkMaterial.CourseID] = append(courseWorkMaterialsMap[courseWorkMaterial.CourseID], courseWorkMaterial)
	}

	// Sets the announcements and courseWorkMaterials of all courses
	for i, course := range newCourses {
		newCourses[i].Announcements = announcementsMap[course.GCID]
		newCourses[i].CourseWorkMaterials = courseWorkMaterialsMap[course.GCID]
	}

	if len(newCourses) != 0 {
		log.Println("Inserting new courses into the database...")
		start := time.Now()
		err = database.SaveCourses(newCourses)
		if err != nil {
			fmt.Println("Error inserting courses into the database:", err)

			http.Error(w, "Failed to insert courses into the database", http.StatusInternalServerError)
			return
		}
		elapsed := time.Since(start)
		log.Printf("Courses successfully inserted into the database in %v", elapsed)
	} else {
		log.Println("No new courses to insert into the database")
	}

	elapsedDiscovery := time.Since(startDiscovery)
	log.Printf("%v courses successfully discovered in %v", len(newCourses), elapsedDiscovery)

	http.Redirect(w, r, os.Getenv("FRONTEND_COURSES_URL"), http.StatusSeeOther)
}

// Retrieves the list of courses for the authenticated user from the database
// Sends them to the client as JSON
func HandleListCourses(w http.ResponseWriter, r *http.Request, store sessions.Store) {
	log.Println("[HandleListCourses] /courses/list hit")
	token, err := database.GetTokenFromSession(r, store)
	if err != nil {
		log.Println("Error retrieving token from the database:", err)
		return
	}

	courses, err := database.GetCoursesByToken(token)
	if err != nil {
		fmt.Println("Error retrieving courses from the database:", err)
		http.Error(w, "Failed to retrieve courses from the database", http.StatusInternalServerError)
		return
	}

	coursesJSON, err := json.Marshal(courses)
	if err != nil {
		fmt.Println("Error marshaling courses to JSON:", err)
		http.Error(w, "Failed to marshal courses to JSON", http.StatusInternalServerError)
		return
	}

	// Set the response content type and write the JSON data to the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(coursesJSON)
}

// Handles request to initiate material download
func HandleDownloadCourses(w http.ResponseWriter, r *http.Request, store sessions.Store) {
	log.Println("[HandleDownloadCourses] /courses/download hit")
	// Parse the request body to get selected courses
	var requestBody struct {
		SelectedCourses []string `json:"selectedCoursesIDs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	token, err := database.GetTokenFromSession(r, store)
	if err != nil {
		log.Println("Error retrieving token from the database:", err)
		return
	}

	err = services.DownloadCourses(requestBody.SelectedCourses, token)
	if err != nil {
		log.Printf("Error during download: %v\n", err)
	}

	w.WriteHeader(http.StatusOK)
}

// Serves the downloaded courses to the client
// Deletes local folders
func HandleServeCourses(w http.ResponseWriter, r *http.Request) {
	log.Println("[HandleServeCourses] /courses/serve hit")
	// Remove the folder that was zipped
	defer os.RemoveAll(utils.DownloadFolderPath)
	// Remove the zip file
	defer os.Remove(utils.ZipFilePath)

	// if utils.DownloadFolderPath is an empty folder, then there is nothing to zip
	// so we return an error
	if isEmpty, err := utils.IsEmptyFolder(utils.DownloadFolderPath); isEmpty || err != nil {
		http.Error(w, "Course Folder is empty.", http.StatusInternalServerError)
		log.Printf("Error checking if folder is empty: %v\n", err)
		return
	}

	// Create a zip file of the download folder
	err := utils.ZipFolder(utils.DownloadFolderPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Open the zip file
	zipFile, err := os.Open(utils.ZipFilePath)
	if err != nil {
		http.Error(w, "Failed to open zip file", http.StatusInternalServerError)
		return
	}
	defer zipFile.Close()

	// Set appropriate headers
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=GCD_"+utils.ZIP_FILE_NAME)

	// Set appropriate headers for cross-origin access
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

	log.Printf("Downloading zip file %s...\n", utils.ZipFilePath)

	// Copy the zip file to the response writer
	_, err = io.Copy(w, zipFile)
	if err != nil {
		http.Error(w, "Failed to copy zip file to response", http.StatusInternalServerError)
		return
	}

	// Close the response writer to ensure all data is flushed
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}
