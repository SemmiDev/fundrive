package fundrive

import (
	"context"
	"golang.org/x/oauth2"
)

type ExchangeTokenRequest struct {
	UserID            string `json:"user_id"`
	AuthorizationCode string `json:"authorization_code"`
}

func (s *ExchangeTokenRequest) Validate() error {
	if s.UserID == "" {
		return ErrInvalidUserID
	}

	if s.AuthorizationCode == "" {
		return ErrInvalidAuthorizationCode
	}

	return nil
}

// ExchangeToken exchanges an authorization code for an OAuth token
func (s *OAuthService) ExchangeToken(ctx context.Context, req *ExchangeTokenRequest) (*oauth2.Token, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	oauth2Token, err := CreateToken(s.OauthConfig, req.AuthorizationCode)
	if err != nil {
		return nil, err
	}

	return oauth2Token, nil
}
