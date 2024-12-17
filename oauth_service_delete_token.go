package fundrive

import (
	"context"
	"fmt"
)

type DeleteTokenRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

func (s *DeleteTokenRequest) Validate() error {
	if s.UserID == "" {
		return ErrInvalidUserID
	}

	if s.Email == "" {
		return ErrInvalidEmail
	}

	return nil
}

// DeleteToken deletes an OAuth token for a user
func (s *OAuthService) DeleteToken(ctx context.Context, req *DeleteTokenRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	err := s.DB.WithContext(ctx).
		Where("user_id = ? AND email = ?", req.UserID, req.Email).
		Delete(&OAuthToken{}).
		Error

	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	return nil
}
