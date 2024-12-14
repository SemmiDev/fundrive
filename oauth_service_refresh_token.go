package fundrive

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
)

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
