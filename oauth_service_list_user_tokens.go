package fundrive

import (
    "context"
    "fmt"
)

type ListUserTokensRequest struct {
    UserID string `json:"user_id"`
}

func (s *ListUserTokensRequest) Validate() error {
    if s.UserID == "" {
        return ErrInvalidUserID
    }

    return nil
}

// ListUserTokens lists all OAuth tokens for a user
func (s *OAuthService) ListUserTokens(ctx context.Context, req *ListUserTokensRequest) ([]OAuthToken, error) {
    if err := req.Validate(); err != nil {
        return nil, err
    }

    tokens := make([]OAuthToken, 0)

    err := s.DB.WithContext(ctx).
        Where("user_id = ?", req.UserID).
        Find(&tokens).
        Error

    if err != nil {
        return nil, fmt.Errorf("failed to list tokens: %w", err)
    }

    return tokens, nil
}
