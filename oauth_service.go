// Package fundrive provides OAuth service functionality
package fundrive

import (
	"context"
	"errors"
	"fmt"

	"github.com/oklog/ulid/v2"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// Domain errors
var (
	ErrTokenNotFound = errors.New("token not found")
	ErrInvalidToken  = errors.New("invalid token provided")
	ErrInvalidUserID = errors.New("invalid user ID provided")
)

// IOAuthService defines the interface for OAuth operations
type IOAuthService interface {
	IsTokenExists(ctx context.Context, userID string) (bool, error)
	GetToken(ctx context.Context, userID string) (*oauth2.Token, error)
	SaveToken(ctx context.Context, userID string, token *oauth2.Token) error
	RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error)
}

// OAuthConfig contains the configuration for OAuth service
type OAuthConfig struct {
	DB           *gorm.DB
	OAuth2Config *oauth2.Config
}

// Validate validates the OAuth configuration
func (c *OAuthConfig) Validate() error {
	if c.DB == nil {
		return errors.New("database connection is required")
	}
	if c.OAuth2Config == nil {
		return errors.New("OAuth2 configuration is required")
	}
	return nil
}

// OAuthService implements IOAuthService interface
type OAuthService struct {
	db          *gorm.DB
	oauthConfig *oauth2.Config
}

// NewOAuthService creates a new instance of OAuthService
func NewOAuthService(config *OAuthConfig) (IOAuthService, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &OAuthService{
		db:          config.DB,
		oauthConfig: config.OAuth2Config,
	}, nil
}

// SaveToken saves or updates an OAuth token for a user
func (s *OAuthService) SaveToken(ctx context.Context, userID string, token *oauth2.Token) error {
	if userID == "" {
		return ErrInvalidUserID
	}
	if token == nil {
		return ErrInvalidToken
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingToken OAuthToken
	err := tx.Where("user_id = ?", userID).First(&existingToken).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newToken := OAuthToken{
			ID:     ulid.Make().String(),
			UserID: userID,
		}
		newToken.FromOAuth2Token(token)

		if err := tx.Create(&newToken).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create token: %w", err)
		}
	} else if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to query token: %w", err)
	} else {
		existingToken.FromOAuth2Token(token)
		if err := tx.Save(&existingToken).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update token: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetToken retrieves an OAuth token for a user
func (s *OAuthService) GetToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}

	var oauthToken OAuthToken
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&oauthToken).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	return oauthToken.ToOAuth2Token(), nil
}

// IsTokenExists checks if a token exists for a user
func (s *OAuthService) IsTokenExists(ctx context.Context, userID string) (bool, error) {
	if userID == "" {
		return false, ErrInvalidUserID
	}

	var count int64
	err := s.db.WithContext(ctx).
		Model(&OAuthToken{}).
		Where("user_id = ?", userID).
		Count(&count).
		Error

	if err != nil {
		return false, fmt.Errorf("failed to check token existence: %w", err)
	}

	return count > 0, nil
}

// RefreshToken refreshes an OAuth token
func (s *OAuthService) RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	if token == nil {
		return nil, ErrInvalidToken
	}

	tokenSource := s.oauthConfig.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return newToken, nil
}
