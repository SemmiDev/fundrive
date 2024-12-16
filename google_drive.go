package fundrive

import (
	"fmt"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// Common errors
var (
	ErrDBEmpty             = fmt.Errorf("error creating google drive service: db is empty")
	ErrInvalidConfig       = fmt.Errorf("invalid configuration provided")
	ErrServiceAccountEmpty = fmt.Errorf("service account file path is empty")
)

// Config represents the configuration for the Google Drive service
type Config struct {
	ServiceAccountFilePath string
	EncryptionKey          string
	DB                     *gorm.DB
}

// ConfigOption defines the function signature for optional configuration
type ConfigOption func(*Config)

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		ServiceAccountFilePath: "",
		EncryptionKey:          "",
		DB:                     nil,
	}
}

// WithServiceAccountFilePath sets the service account file path
func WithServiceAccountFilePath(path string) ConfigOption {
	return func(c *Config) {
		c.ServiceAccountFilePath = path
	}
}

// WithDB sets the database connection
func WithDB(db *gorm.DB) ConfigOption {
	return func(c *Config) {
		c.DB = db
	}
}

// WithEncryptionKey sets the encryption key
func WithEncryptionKey(key string) ConfigOption {
	return func(c *Config) {
		c.EncryptionKey = key
	}
}

// validate checks if the configuration is valid
func (c *Config) validate() error {
	if c.DB == nil {
		return ErrDBEmpty
	}
	if c.ServiceAccountFilePath == "" {
		return ErrServiceAccountEmpty
	}

	if c.EncryptionKey == "" {
		return ErrServiceAccountEmpty
	}

	return nil
}

type GoogleDriveService struct {
	OAuthService IOAuthService
	OauthConfig  *oauth2.Config
}

// New creates a new instance of IGoogleDriveService with the provided configuration
func New(opts ...ConfigOption) (*GoogleDriveService, error) {
	// Start with default configuration
	config := DefaultConfig()

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
	}

	return &service, nil
}
