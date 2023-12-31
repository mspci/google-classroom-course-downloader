package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mspcix/google-classroom-course-downloader/database"
	"github.com/mspcix/google-classroom-course-downloader/models"
	"github.com/mspcix/google-classroom-course-downloader/utils"
	"golang.org/x/oauth2"
)

func PopulateUserProfile(ctx context.Context, token *oauth2.Token) (*models.User, error) {
	httpClient := utils.OAuthConfig.Client(ctx, token)

	url := "https://classroom.googleapis.com/v1/userProfiles/me"
	response, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting user profile: %w", err)
	}
	defer response.Body.Close()

	var userGCProfile models.GCProfile
	err = json.NewDecoder(response.Body).Decode(&userGCProfile)
	if err != nil {
		return nil, fmt.Errorf("error decoding response body: %w", err)
	}

	user := models.User{
		GCUID:        userGCProfile.GClassroomID,
		Username:     userGCProfile.Name.FullName,
		Email:        userGCProfile.EmailAddress,
		Token:        token.AccessToken,
		TokenExpiry:  token.Expiry,
		RefreshToken: token.RefreshToken,
		PhotoUrl:     userGCProfile.PhotoUrl,
	}

	return &user, nil
}

// RefreshToken refreshes the access token using the refresh token
func RefreshToken(expiredToken *string) {
	// Get the refresh token from the database
	refreshToken, err := database.GetRefreshTokenByToken(*expiredToken)
	if err != nil {
		log.Println("Error getting refresh token:", err)
		return
	}

	token := &oauth2.Token{RefreshToken: refreshToken}
	newToken, err := utils.OAuthConfig.TokenSource(context.Background(), token).Token()
	if err != nil {
		log.Println("Error getting new token:", err)
		return
	}

	err = database.UpdateToken(*expiredToken, newToken)
	if err != nil {
		log.Println("Error updating token:", err)
		return
	}

	*expiredToken = newToken.AccessToken
	log.Println("Token refreshed")
}
