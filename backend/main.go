package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/rs/cors"

	"github.com/mspcix/google-classroom-course-downloader/database"
	"github.com/mspcix/google-classroom-course-downloader/routes"
	"github.com/mspcix/google-classroom-course-downloader/utils"
)

func main() {
	err := utils.InitEnv()
	if err != nil {
		log.Fatal("Error initializing the environment:", err)
		return
	}

	utils.InitLogger()
	defer utils.Logger.Writer().(*os.File).Close()
	defer utils.DBLogger.LogFile.Close()

	err = utils.InitOauthConfig()
	if err != nil {
		log.Fatal("Error initializing the OAuth config:", err)
		return
	}

	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Error initializing the database:", err)
		return
	}

	pgDB, err := db.DB()
	if err != nil {
		log.Fatal("Error getting the database connection:", err)
		return
	}
	defer pgDB.Close()
	defer db.Exec("CLOSE ALL")

	r := mux.NewRouter()

	cookieStore := sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	cookieStore.Options.MaxAge = 60 * 60 * 24 // 1 day in seconds

	// Apply CORS middleware to the router
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{os.Getenv("FRONTEND_URL"), os.Getenv("SERVER_URL"), os.Getenv("FRONTEND_COURSES_URL")},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Origin",
			"Accept",
			"Content-Type",
			"Authorization",
		},
		AllowCredentials: true,
		// Debug:            true,
	})

	// Use the corsMiddleware as a handler
	http.Handle("/", corsMiddleware.Handler(r))

	routes.SetupRoutes(r, cookieStore)

	fmt.Println("Server started at " + os.Getenv("SERVER_URL"))

	http.ListenAndServe(":"+os.Getenv("SERVER_PORT"), nil)
}
