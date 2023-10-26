package database

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/mspcix/google-classroom-course-downloader/models"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

func SaveUser(user models.User) error {
	existingUser, err := GetUserByGCUID(user.GCUID)
	if err != nil {
		return fmt.Errorf("error retrieving user from the database: %w", err)
	}

	if existingUser == nil {
		log.Println("User does not exist in the database. Inserting...")
		return insertUser(user)
	}

	if user.Token != existingUser.Token {
		log.Println("User exists in the database. Updating user's token...")
		return UpdateUserToken(user)
	}

	return nil
}

func insertUser(user models.User) error {
	result := db.Create(&user)
	if result.Error != nil {
		return fmt.Errorf("error inserting user into the database: %w", result.Error)
	}
	return nil
}

func UpdateUserToken(user models.User) error {
	result := db.Model(&models.User{}).Where("email = ?", user.Email).
		Updates(map[string]interface{}{
			"token":         user.Token,
			"refresh_token": user.RefreshToken,
			"updated_at":    time.Now(),
		})
	if result.Error != nil {
		return fmt.Errorf("error updating user in the database: %w", result.Error)
	}
	return nil
}

func GetUserByGCUID(gcuid string) (*models.User, error) {
	var user models.User

	// Check if the "users" table exists before querying
	migrator := db.Migrator()
	if migrator.HasTable(&user) {
		result := db.Where("gc_user_id = ?", gcuid).First(&user)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return nil, nil // User does not exist
			}
			return nil, fmt.Errorf("error retrieving user from the database: %w", result.Error)
		}
		return &user, nil
	}

	// If the "users" table does not exist, return nil
	return nil, fmt.Errorf("THE 'USERS' TABLE DOESN'T EXIST IN THE DATABASE. PlEASE, RESTART THE SERVER")
}

// Get user's token from session data
func GetTokenFromSession(r *http.Request, store sessions.Store) (string, error) {
	// Retrieve the authenticated user's userID from the session
	session, _ := store.Get(r, "gcd_session")
	gcuid, ok := session.Values["gcuid"].(string)
	if !ok {
		log.Println("userID not found in session")
		return "", fmt.Errorf("userID not found in session")
	}

	token, err := GetTokenByGCUID(gcuid)
	if err != nil {
		log.Println("Error retrieving token from the database:", err)
		return "", err
	}

	return token, nil
}

func GetTokenByGCUID(gcuid string) (string, error) {
	var user models.User

	migrator := db.Migrator()
	if migrator.HasTable(&user) {
		result := db.Select("token").Where("gc_user_id = ?", gcuid).First(&user)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return "", nil // User does not exist
			}
			return "", fmt.Errorf("error retrieving user from the database: %w", result.Error)
		}
		return user.Token, nil
	}

	return "", fmt.Errorf("THE 'USERS' TABLE DOESN'T EXIST IN THE DATABASE")
}

func GetGCUIDByToken(token string) (string, error) {
	var user models.User

	migrator := db.Migrator()
	if migrator.HasTable(&user) {
		result := db.Select("gc_user_id").Where("token = ?", token).First(&user)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return "", nil // User does not exist
			}
			return "", fmt.Errorf("error retrieving user from the database: %w", result.Error)
		}
		return user.GCUID, nil
	}

	return "", fmt.Errorf("THE 'USERS' TABLE DOESN'T EXIST IN THE DATABASE")
}

func GetRefreshTokenByToken(token string) (string, error) {
	var user models.User

	migrator := db.Migrator()
	if migrator.HasTable(&user) {
		result := db.Select("refresh_token").Where("token = ?", token).First(&user)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				return "", nil // User does not exist
			}
			return "", fmt.Errorf("error retrieving user from the database: %w", result.Error)
		}
		return user.RefreshToken, nil
	}

	return "", fmt.Errorf("THE 'USERS' TABLE DOESN'T EXIST IN THE DATABASE")
}

func UpdateToken(expiredToken string, newToken *oauth2.Token) error {
	result := db.Model(&models.User{}).Where("token = ?", expiredToken).
		Updates(map[string]interface{}{
			"token":         newToken.AccessToken,
			"token_expiry":  newToken.Expiry,
			"refresh_token": newToken.RefreshToken,
			"updated_at":    time.Now(),
		})
	if result.Error != nil {
		return fmt.Errorf("error updating user in the database: %w", result.Error)
	}
	return nil
}
