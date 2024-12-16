package fundrive

import (
	"fmt"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type GoogleDriveService struct {
	OAuthService IOAuthService
	OauthConfig  *oauth2.Config
	DB           *gorm.DB
}

// New creates a new instance of IGoogleDriveService with the provided configuration
func New(opts ...GoogleDriveServiceConfigOption) (*GoogleDriveService, error) {
	// Start with default configuration
	config := DefaultGoogleDriveServiceConfig()

	// Apply all provided options
	for _, opt := range opts {
		opt(config)
	}

	// Validate configuration
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Auto migrate database schema
	if err := config.DB.AutoMigrate(&OAuthToken{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Initialize OAuth2 configuration
	oauth2Config, err := NewOAuth2Config(config.ServiceAccountFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth config: %w", err)
	}

	// Initialize token encryption
	tokenEncryptor, err := NewTokenEncryption(config.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create token encryption: %w", err)
	}

	// Initialize OAuth config
	oauthConfig := OAuthConfig{
		DB:             config.DB,
		OAuth2Config:   oauth2Config,
		TokenEncryptor: tokenEncryptor,
	}

	// Initialize OAuth service
	oauthService, err := NewOAuthService(&oauthConfig)
	if err != nil {
		return nil, err
	}

	// Create service instance
	service := GoogleDriveService{
		OAuthService: oauthService,
		OauthConfig:  oauth2Config,
		DB:           config.DB,
	}

	return &service, nil
}
