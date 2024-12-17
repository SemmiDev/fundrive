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
	UserID          string        `json:"user_id"`
	Email           string        `json:"email"`
	BaseFolderID    *string       `json:"base_folder_id"`
	ExpiryTimestamp *string       `json:"expiry_timestamp"`
	Token           *oauth2.Token `json:"token"`
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

	tx := s.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	var existingToken OAuthToken

	err := tx.
		Where("user_id = ? AND email = ?", req.UserID, req.Email).
		First(&existingToken).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newOauthToken := OAuthToken{
			ID:              ulid.Make().String(),
			UserID:          req.UserID,
			Email:           req.Email,
			BaseFolderID:    req.BaseFolderID,
			ExpiryTimestamp: req.ExpiryTimestamp,
		}

		if err := newOauthToken.FromOAuth2Token(req.Token, s.TokenEncryptor); err != nil {
			return fmt.Errorf("failed to convert token: %w", err)
		}

		if err := tx.Create(&newOauthToken).Error; err != nil {
			return fmt.Errorf("failed to create token: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to query token: %w", err)
	} else {
		if err := existingToken.FromOAuth2Token(req.Token, s.TokenEncryptor); err != nil {
			return fmt.Errorf("failed to convert token: %w", err)
		}

		if err := tx.Save(&existingToken).Error; err != nil {
			return fmt.Errorf("failed to update token: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
