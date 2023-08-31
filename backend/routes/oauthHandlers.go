package routes

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"golang.org/x/oauth2"

	"github.com/mspcix/google-classroom-course-downloader/database"
	"github.com/mspcix/google-classroom-course-downloader/services"
	"github.com/mspcix/google-classroom-course-downloader/utils"
)

var storedState string

// Generates the authentication URL and sends it to the frontend.
func HandleOAuthURL(w http.ResponseWriter, r *http.Request, store sessions.Store) {
	log.Println("oauth/url route hit")

	// Generate random state for oauth2 flow (protecting against CSRF)
	storedState = utils.GenerateRandomID(32)

	url := utils.OAuthConfig.AuthCodeURL(storedState, oauth2.AccessTypeOffline)
	fmt.Fprintf(w, `{"url": "%s"}`, url)
}

func HandleOAuthCallback(w http.ResponseWriter, r *http.Request, store sessions.Store) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if storedState != state {
		log.Println("Invalid state parameter")
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}
	storedState = ""

	token, err := utils.OAuthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Println("Error exchanging code for token:", err)
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	user, err := services.PopulateUserProfile(r.Context(), token)
	if err != nil {
		log.Println("Error getting user profile:", err)
		http.Error(w, "Error getting user profile", http.StatusInternalServerError)
		return
	}

	err = database.SaveUser(*user)
	if err != nil {
		log.Println("Error saving user to the database:", err)
		http.Error(w, "Error saving user to the database", http.StatusInternalServerError)
		return
	}

	session, _ := store.Get(r, "gcd_session")
	session.Values["authenticated"] = true
	session.Values["gcuid"] = user.GCUID
	session.Save(r, w)

	http.Redirect(w, r, "/courses/discover", http.StatusSeeOther)
}

func HandleLogout(w http.ResponseWriter, r *http.Request, store sessions.Store) {
	log.Println("[HandleLogout] /logout hit")

	session, _ := store.Get(r, "gcd_session")

	// Delete the session data
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
