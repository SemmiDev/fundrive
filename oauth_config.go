package fundrive

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"os"
	"strings"
)

var (
	ErrServiceAccountPathEmpty = fmt.Errorf("error set up oauth config: please provide service account path")
	ErrServiceAccountFileEmpty = fmt.Errorf("error set up oauth config: service account file is empty")
	ErrClientIDEmpty           = fmt.Errorf("error set up oauth config: client ID is empty")
	ErrClientSecretEmpty       = fmt.Errorf("error set up oauth config: client secret is empty")
)

// userScopes defines the OAuth scopes required for user information
var userScopes = []string{
	"https://www.googleapis.com/auth/userinfo.email",
	"https://www.googleapis.com/auth/userinfo.profile",
}

// driveScopes defines the OAuth scopes required for Google Drive access
var driveScopes = []string{
	drive.DriveScope,
	drive.DriveReadonlyScope,
	drive.DriveMetadataReadonlyScope,
	drive.DriveMetadataScope,
	drive.DriveFileScope,
	drive.DriveScriptsScope,
}

// createConfig creates an OAuth2 config from service account JSON data
func createConfig(data []byte) (*oauth2.Config, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("error read service account file: %w", ErrServiceAccountFileEmpty)
	}

	scopes := append(userScopes, driveScopes...)
	oauth2Config, err := google.ConfigFromJSON(data, scopes...)
	if err != nil {
		return nil, fmt.Errorf("error set up oauth config from json: %w", err)
	}

	return oauth2Config, nil
}

// NewOAuth2Config creates a new OAuth2 config from a service account file path
func NewOAuth2Config(serviceAccountPath string) (*oauth2.Config, error) {
	if strings.TrimSpace(serviceAccountPath) == "" {
		return nil, fmt.Errorf("error read service account path: %w", ErrServiceAccountPathEmpty)
	}

	data, err := os.ReadFile(serviceAccountPath)
	if err != nil {
		return nil, fmt.Errorf("error read service account file: %w", err)
	}

	return createConfig(data)
}

// OAuthCredentials contains the credentials needed for OAuth2 configuration
type OAuthCredentials struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// NewOAuth2ConfigFromCredentials creates a new OAuth2 config using manual credentials
func NewOAuth2ConfigFromCredentials(cred OAuthCredentials) (*oauth2.Config, error) {
	if cred.ClientID == "" {
		return nil, fmt.Errorf("error creating oauth config: %w", ErrClientIDEmpty)
	}
	if cred.ClientSecret == "" {
		return nil, fmt.Errorf("error creating oauth config: %w", ErrClientSecretEmpty)
	}

	scopes := append(userScopes, driveScopes...)

	config := &oauth2.Config{
		ClientID:     cred.ClientID,
		ClientSecret: cred.ClientSecret,
		RedirectURL:  cred.RedirectURL,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}

	return config, nil
}

// MustNewOAuth2ConfigFromCredentials creates a new OAuth2 config using credentials or panics
func MustNewOAuth2ConfigFromCredentials(cred OAuthCredentials) *oauth2.Config {
	config, err := NewOAuth2ConfigFromCredentials(cred)
	if err != nil {
		panic(err)
	}
	return config
}

// MustNewOAuthConfig creates a new OAuth2 config from a service account file path or panics
func MustNewOAuthConfig(serviceAccountPath string) *oauth2.Config {
	config, err := NewOAuth2Config(serviceAccountPath)
	if err != nil {
		panic(err)
	}
	return config
}

// NewOAuthConfigFromFile creates a new OAuth2 config from service account file bytes
func NewOAuthConfigFromFile(serviceAccountFile []byte) (*oauth2.Config, error) {
	if serviceAccountFile == nil {
		return nil, fmt.Errorf("error read service account file: %w", ErrServiceAccountPathEmpty)
	}
	return createConfig(serviceAccountFile)
}

// MustNewOAuthConfigFromFile creates a new OAuth2 config from service account file bytes or panics
func MustNewOAuthConfigFromFile(serviceAccountFile []byte) *oauth2.Config {
	config, err := NewOAuthConfigFromFile(serviceAccountFile)
	if err != nil {
		panic(err)
	}
	return config
}

// CreateToken creates an OAuth2 token using credentials and authorization code
func CreateToken(config *oauth2.Config, authCode string) (*oauth2.Token, error) {
	token, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		return nil, fmt.Errorf("error exchanging auth code for token: %w", err)
	}
	return token, nil
}
