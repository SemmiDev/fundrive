package fundrive

import (
	"fmt"
	"gorm.io/gorm"
)

// Common errors
var (
	ErrDBEmpty             = fmt.Errorf("error creating google drive service: db is empty")
	ErrInvalidConfig       = fmt.Errorf("invalid configuration provided")
	ErrServiceAccountEmpty = fmt.Errorf("service account file path is empty")
)

// GoogleDriveServiceConfig represents the configuration for the Google Drive service
type GoogleDriveServiceConfig struct {
	ServiceAccountFilePath string
	EncryptionKey          string
	DB                     *gorm.DB
	UseBaseFolder          bool
}

// GoogleDriveServiceConfigOption defines the function signature for optional configuration
type GoogleDriveServiceConfigOption func(*GoogleDriveServiceConfig)

// DefaultGoogleDriveServiceConfig returns a Config with default values
func DefaultGoogleDriveServiceConfig() *GoogleDriveServiceConfig {
	return &GoogleDriveServiceConfig{}
}

// WithServiceAccountFilePath sets the service account file path
func WithServiceAccountFilePath(path string) GoogleDriveServiceConfigOption {
	return func(c *GoogleDriveServiceConfig) {
		c.ServiceAccountFilePath = path
	}
}

// WithDB sets the database connection
func WithDB(db *gorm.DB) GoogleDriveServiceConfigOption {
	return func(c *GoogleDriveServiceConfig) {
		c.DB = db
	}
}

// WithEncryptionKey sets the encryption key
func WithEncryptionKey(key string) GoogleDriveServiceConfigOption {
	return func(c *GoogleDriveServiceConfig) {
		c.EncryptionKey = key
	}
}

func WithUseBaseFolder(useBaseFolder bool) GoogleDriveServiceConfigOption {
	return func(c *GoogleDriveServiceConfig) {
		c.UseBaseFolder = useBaseFolder
	}
}

// validate checks if the configuration is valid
func (c *GoogleDriveServiceConfig) validate() error {
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
