package routes

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type HandlerWStore func(http.ResponseWriter, *http.Request, sessions.Store)

func SetupRoutes(r *mux.Router, store sessions.Store) {
	r.HandleFunc("/", withStore(HandleHome, store))
	r.HandleFunc("/oauth/url", withStore(HandleOAuthURL, store))
	r.HandleFunc("/oauth/callback", withStore(HandleOAuthCallback, store))
	r.HandleFunc("/courses/discover", authMiddleware(withStore(HandleDiscoverCourses, store), store))
	r.HandleFunc("/courses/list", authMiddleware(withStore(HandleListCourses, store), store))
	r.HandleFunc("/courses/download", authMiddleware(withStore(HandleDownloadCourses, store), store))
	r.HandleFunc("/courses/serve", authMiddleware(HandleServeCourses, store))
}

// Checks if the user is authenticated
// and has a valid session before allowing access to protected routes.
func authMiddleware(next http.HandlerFunc, store sessions.Store) http.HandlerFunc {
	log.Println("authMiddleware route hit")
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "GCD_session")
		// Check if the user is authenticated
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			log.Println("user is not authenticated (middleware function)")
			w.WriteHeader(http.StatusUnauthorized)
			// http.Redirect(w, r, "/oauth/url", http.StatusSeeOther)
			// http.Redirect(w, r, os.Getenv("FRONTEND_URL"), http.StatusSeeOther)
			return
		}
		// Call the next handler if authenticated
		next.ServeHTTP(w, r)
	}
}

func withStore(h HandlerWStore, store sessions.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h(w, r, store)
	}
}

func HandleHome(w http.ResponseWriter, r *http.Request, store sessions.Store) {
	log.Println("home route hit")

	// Check if the user has a session
	session, _ := store.Get(r, "GCD_session")

	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		// http.Redirect(w, r, "/oauth/url", http.StatusSeeOther)
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		log.Println("user is authenticated")
		// http.Redirect(w, r, "/courses/discover", http.StatusSeeOther)
		w.WriteHeader(http.StatusOK)
	}
}
