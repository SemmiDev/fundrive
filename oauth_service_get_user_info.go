package fundrive

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"io"
	"net/http"
)

// GoogleUserInfo represents basic user information from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// GetUserInfoRequest represents the request for getting user information
type GetUserInfoRequest struct {
	Token *oauth2.Token `json:"token"`
}

// GetGoogleUserInfo retrieves the user's information from Google
func (s *OAuthService) GetGoogleUserInfo(ctx context.Context, req *GetUserInfoRequest) (*GoogleUserInfo, error) {
	if req == nil || req.Token == nil || req.Token.AccessToken == "" {
		return nil, ErrInvalidToken
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + req.Token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Check response status
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get user info: status code %d, body %s", resp.StatusCode, resp.Body)
	}

	userData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Parse response into struct
	var result GoogleUserInfo
	if err := json.Unmarshal(userData, &result); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	// Validate essential fields
	if result.Email == "" {
		return nil, ErrInvalidEmail
	}

	if result.ID == "" {
		return nil, ErrInvalidUserID
	}

	return &result, nil
}
