package fundrive

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type GetTokenRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

func (s *GetTokenRequest) Validate() error {
	if s.UserID == "" {
		return ErrInvalidUserID
	}

	if s.Email == "" {
		return ErrInvalidEmail
	}

	return nil
}

// GetToken retrieves an OAuth token for a user
func (s *OAuthService) GetToken(ctx context.Context, req *GetTokenRequest) (*oauth2.Token, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var oauthToken OAuthToken
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND email = ?", req.UserID, req.Email).
		First(&oauthToken).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	oauth2Token, err := oauthToken.ToOAuth2Token(s.TokenEncryptor)
	if err != nil {
		return nil, err
	}

	return oauth2Token, nil
}
