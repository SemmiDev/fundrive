package fundrive

import (
	"context"
	"errors"
	"fmt"

	"github.com/oklog/ulid/v2"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type SaveTokenRequest struct {
	UserID string        `json:"user_id"`
	Email  string        `json:"email"`
	Token  *oauth2.Token `json:"token"`
}

func (s *SaveTokenRequest) Validate() error {
	if s.UserID == "" {
		return ErrInvalidUserID
	}

	if s.Email == "" {
		return ErrInvalidEmail
	}

	if s.Token == nil {
		return ErrInvalidToken
	}

	return nil
}

func (s *OAuthService) SaveToken(ctx context.Context, req *SaveTokenRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	tx := s.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	var existingToken OAuthToken
	err := tx.
		Where("user_id = ? AND email = ?", req.UserID, req.Email).
		First(&existingToken).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newToken := OAuthToken{
			ID:     ulid.Make().String(),
			UserID: req.UserID,
			Email:  req.Email,
		}

		newToken.FromOAuth2Token(req.Token)

		if err := tx.Create(&newToken).Error; err != nil {
			return fmt.Errorf("failed to create token: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to query token: %w", err)
	} else {
		existingToken.FromOAuth2Token(req.Token)
		if err := tx.Save(&existingToken).Error; err != nil {
			return fmt.Errorf("failed to update token: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
