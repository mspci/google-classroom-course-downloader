package utils

import (
	"archive/zip"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/option"
	"gorm.io/gorm/logger"
)

type GormLogger struct {
	LoggerInterface logger.Interface
	LogFile         *os.File
}

var (
	SystemDownloadFolder  string
	DownloadFolder        string
	ZIP_FILE_NAME         string
	OAuthConfig           *oauth2.Config
	DownloadFolderPath, _ string
	ZipFilePath           string
	Logger                *log.Logger
	DBLogger              GormLogger
)

func InitEnv() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	SystemDownloadFolder = os.Getenv("SYSTEM_DOWNLOAD_FOLDER")
	DownloadFolder = os.Getenv("DOWNLOAD_FOLDER")

	DownloadFolderPath = DefineDownloadPath()
	ZIP_FILE_NAME = DownloadFolder + ".zip"
	ZipFilePath = DownloadFolderPath + ".zip"

	log.Println("------------------------------------------------------")
	log.Println("------------------------------------------------------")
	log.Printf("SystemDownloadFolder: %s\n", SystemDownloadFolder)
	log.Printf("DownloadFolder: %s\n", DownloadFolder)
	log.Printf("DownloadFolderPath: %s\n", DownloadFolderPath)
	log.Printf("ZIP_FILE_NAME: %s\n", ZIP_FILE_NAME)
	log.Printf("ZipFilePath: %s\n", ZipFilePath)
	log.Println("------------------------------------------------------")
	log.Println("------------------------------------------------------")

	return nil
}

func InitOauthConfig() error {
	OAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("SERVER_URL") + os.Getenv("ROUTE_OAUTH_CALLBACK"),
		Scopes: []string{`https://www.googleapis.com/auth/classroom.courses.readonly`,
			`https://www.googleapis.com/auth/classroom.announcements.readonly`,
			`https://www.googleapis.com/auth/classroom.coursework.me.readonly`,
			`https://www.googleapis.com/auth/classroom.courseworkmaterials.readonly`,
			`https://www.googleapis.com/auth/classroom.guardianlinks.me.readonly`,
			`https://www.googleapis.com/auth/classroom.profile.emails`,
			`https://www.googleapis.com/auth/classroom.profile.photos`,
			`https://www.googleapis.com/auth/classroom.push-notifications`,
			`https://www.googleapis.com/auth/classroom.rosters.readonly`,
			`https://www.googleapis.com/auth/classroom.student-submissions.me.readonly`,
			`https://www.googleapis.com/auth/classroom.topics.readonly`,
			`https://www.googleapis.com/auth/drive.file`,
			`https://www.googleapis.com/auth/drive.readonly`,
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		},
	}

	return nil
}

func RemoveInvalidChars(fileName string) string {
	// Define a regular expression pattern to match invalid characters
	invalidCharPattern := regexp.MustCompile(`[<>:"'/\\|?*]`)

	// Replace invalid characters with an underscore
	sanitizedFileName := invalidCharPattern.ReplaceAllString(fileName, "_")

	// Also trim leading and trailing whitespace and dots
	sanitizedFileName = strings.Trim(sanitizedFileName, " .")

	return sanitizedFileName
}

func DownloadDriveFile(token, fileID, filePath string) error {
	ctx := context.Background()

	// Set up the Drive API client
	client := getClient(ctx, token)

	// Download the file content
	resp, err := client.Files.Get(fileID).Download()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the local file
	localFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	// Copy the downloaded content to the local file
	_, err = io.Copy(localFile, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func getClient(ctx context.Context, token string) *drive.Service {
	svc, err := drive.NewService(ctx, option.WithTokenSource(OAuthConfig.TokenSource(ctx, &oauth2.Token{AccessToken: token})))
	if err != nil {
		log.Fatalf("Unable to create Drive service: %v", err)
	}
	return svc
}

// Convert an RFC3339 time string to a folder name
func MakeFolderNameFromTime(creationTime string) string {
	// Parse the creation date string into a time.Time object
	folderName, err := time.Parse(time.RFC3339, creationTime)
	if err != nil {
		log.Fatalf("Unable to parse creation time: %v", err)
	}

	// Convert creation time to a format suitable for a folder name
	creationDate := folderName.Format("02-01-2006")

	return creationDate
}

// Create a zip file from a folder
func createZip(sourceDir, zipFilePath string) error {
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return filepath.Walk(sourceDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relPath, err := filepath.Rel(sourceDir, filePath)
			if err != nil {
				return err
			}

			zipFile, err := zipWriter.Create(relPath)
			if err != nil {
				return err
			}

			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(zipFile, file)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func ZipFolder(sourceDir string) error {
	// Create a zip file with the same name as the folder
	ZipFilePath = sourceDir + ".zip"
	log.Printf("Zipping folder %s...\n", ZipFilePath)

	// Close any open file handles within the directory
	if err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
		}
		return nil
	}); err != nil {
		return err
	}

	// Create the zip file
	err := createZip(sourceDir, ZipFilePath)
	if err != nil {
		return err
	}

	return nil
}

func DefineDownloadPath() string {
	SystemDownloadFolder = os.Getenv("SYSTEM_DOWNLOAD_FOLDER")
	DownloadFolder = os.Getenv("DOWNLOAD_FOLDER")

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Unable to get user's home directory:", err)
	}

	downloadPath := filepath.Join(userHomeDir, SystemDownloadFolder, DownloadFolder)

	return downloadPath
}

// Generates a random session ID
// Returns a unique session identifier
func GenerateRandomID(IDlength int) string {
	// Generate a random 8-character string
	b := make([]byte, IDlength)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatalf("Unable to generate session ID: %v", err)
	}

	// Encode the random bytes as a base64 string
	return base64.StdEncoding.EncodeToString(b)
}

func IsEmptyFolder(folderPath string) (bool, error) {
	// Open the folder
	folder, err := os.Open(folderPath)
	if err != nil {
		return false, err
	}
	defer folder.Close()

	// Read in the files
	files, err := folder.Readdir(0)
	if err != nil {
		return false, err
	}

	// Return true if the folder is empty
	return len(files) == 0, nil
}

func InitLogger() {
	logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	Logger = log.New(logFile, "", log.LstdFlags)

	dbLogFile, err := os.OpenFile("gorm.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	customGormLogger := logger.New(
		log.New(dbLogFile, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,  // Slow SQL threshold
			LogLevel:                  logger.Error, // Log level
			IgnoreRecordNotFoundError: false,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,        // Disable color
		},
	)

	// Populate a GormLogger struct
	DBLogger = GormLogger{
		LoggerInterface: customGormLogger,
		LogFile:         dbLogFile,
	}
}

func GetGCUIDFromSession(r *http.Request, store sessions.Store) (string, error) {
	session, _ := store.Get(r, "gcd_session")

	gcuid, ok := session.Values["gcuid"].(string)
	if !ok {
		return "", fmt.Errorf("unable to get GCUID from session")
	}

	return gcuid, nil
}
