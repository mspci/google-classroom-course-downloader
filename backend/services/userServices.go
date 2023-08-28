package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mspcix/google-classroom-downloader/models"
	"github.com/mspcix/google-classroom-downloader/utils"
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
		GCUID:    userGCProfile.GClassroomID,
		Username: userGCProfile.Name.FullName,
		Email:    userGCProfile.EmailAddress,
		Token:    token.AccessToken,
		PhotoUrl: userGCProfile.PhotoUrl,
	}

	return &user, nil
}
