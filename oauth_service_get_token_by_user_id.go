package fundrive

import (
    "context"
    "errors"
    "fmt"
    "gorm.io/gorm"
)

type GetTokenByUserIDRequest struct {
    UserID string `json:"user_id"`
}

func (s *GetTokenByUserIDRequest) Validate() error {
    if s.UserID == "" {
        return ErrInvalidUserID
    }

    return nil
}

// GetTokenByUserID get first OAuth tokens for a user
func (s *OAuthService) GetTokenByUserID(ctx context.Context, req *GetTokenByUserIDRequest) (*OAuthToken, error) {
    if err := req.Validate(); err != nil {
        return nil, err
    }

    var token OAuthToken

    err := s.DB.WithContext(ctx).
        Where("user_id = ?", req.UserID).
        First(&token).
        Error

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrTokenNotFound
        }
        return nil, fmt.Errorf("failed to get token by user id: %w", err)
    }

    return &token, nil
}
