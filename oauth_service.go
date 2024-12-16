package fundrive

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// Domain errors
var (
	ErrTokenNotFound = errors.New("token not found")
	ErrInvalidToken  = errors.New("invalid token provided")
	ErrInvalidUserID = errors.New("invalid user ID provided")
	ErrInvalidEmail  = errors.New("invalid email provided")
)

// IOAuthService defines the interface for OAuth operations
type IOAuthService interface {
	GetToken(ctx context.Context, req *GetTokenRequest) (*oauth2.Token, error)
	SaveToken(ctx context.Context, req *SaveTokenRequest) error
	IsTokenExists(ctx context.Context, req *IsTokenExistsRequest) (bool, error)
	RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error)
	GetGoogleUserInfo(ctx context.Context, req *GetUserInfoRequest) (*GoogleUserInfo, error)
}

// OAuthConfig contains the configuration for OAuth service
type OAuthConfig struct {
	DB             *gorm.DB
	OAuth2Config   *oauth2.Config
	TokenEncryptor *TokenEncryption
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
	db             *gorm.DB
	oauthConfig    *oauth2.Config
	TokenEncryptor *TokenEncryption
}

// NewOAuthService creates a new instance of OAuthService
func NewOAuthService(config *OAuthConfig) (IOAuthService, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &OAuthService{
		db:             config.DB,
		oauthConfig:    config.OAuth2Config,
		TokenEncryptor: config.TokenEncryptor,
	}, nil
}
