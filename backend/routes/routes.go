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
	r.HandleFunc("/oauth/logout", withStore(HandleLogout, store))
	r.HandleFunc("/oauth/callback", withStore(HandleOAuthCallback, store))
	r.HandleFunc("/courses/discover", withStore(HandleDiscoverCourses, store))
	r.HandleFunc("/courses/list", withStore(HandleListCourses, store))
	r.HandleFunc("/courses/download", withStore(HandleDownloadCourses, store))
	r.HandleFunc("/courses/serve", HandleServeCourses)
}

// Checks if the user is authenticated
// and has a valid session before allowing access to protected routes.
func authMiddleware(next http.HandlerFunc, store sessions.Store) http.HandlerFunc {
	log.Println("authMiddleware route hit")
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "gcd_session")
		// Check if the user is authenticated
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			log.Println("user is not authenticated (middleware function)")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

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
	session, _ := store.Get(r, "gcd_session")

	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
