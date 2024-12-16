package fundrive

import (
	"context"
	"fmt"
)

type IsTokenExistsRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

func (s *IsTokenExistsRequest) Validate() error {
	if s.UserID == "" {
		return ErrInvalidUserID
	}

	if s.Email == "" {
		return ErrInvalidEmail
	}

	return nil
}

// IsTokenExists checks if a token exists for a user
func (s *OAuthService) IsTokenExists(ctx context.Context, req *IsTokenExistsRequest) (bool, error) {
	if err := req.Validate(); err != nil {
		return false, err
	}

	var count int64
	err := s.DB.WithContext(ctx).
		Model(&OAuthToken{}).
		Where("user_id = ? AND email = ?", req.UserID, req.Email).
		Count(&count).
		Error

	if err != nil {
		return false, fmt.Errorf("failed to check token existence: %w", err)
	}

	return count > 0, nil
}
