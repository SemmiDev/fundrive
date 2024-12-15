// Package fundrive handles Google Drive OAuth2 configuration
package fundrive

import (
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
