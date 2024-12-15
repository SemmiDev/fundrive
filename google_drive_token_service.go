package fundrive

import (
	"context"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type newDriveServiceRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

func (service *GoogleDriveService) newDriveService(ctx context.Context, req *newDriveServiceRequest) (*drive.Service, error) {

	getTokenReq := GetTokenRequest{
		UserID: req.UserID,
		Email:  req.Email,
	}

	token, err := service.OAuthService.GetToken(ctx, &getTokenReq)
	if err != nil {
		return nil, err
	}

	newTokenServiceReq := newTokenServiceRequest{
		UserID: req.UserID,
		Email:  req.Email,
		Token:  token,
	}

	srv, err := service.newTokenService(ctx, &newTokenServiceReq)
	if err != nil {
		return nil, err
	}

	return srv, nil
}

type newTokenServiceRequest struct {
	UserID string        `json:"user_id"`
	Email  string        `json:"email"`
	Token  *oauth2.Token `json:"token"`
}

// newTokenService creates a new Google Drive service using the provided token.
// If the token is invalid, it will refresh the token and create a new service.
func (service *GoogleDriveService) newTokenService(
	ctx context.Context,
	req *newTokenServiceRequest,
) (*drive.Service, error) {

	if !req.Token.Valid() {
		refreshedToken, err := service.OAuthService.RefreshToken(ctx, req.Token)
		if err != nil {
			return nil, err
		}

		req.Token = refreshedToken

		saveTokenReq := SaveTokenRequest{
			UserID: req.UserID,
			Email:  req.Email,
			Token:  req.Token,
		}

		if err = service.OAuthService.SaveToken(ctx, &saveTokenReq); err != nil {
			return nil, err
		}
	}

	tokenSource := oauth2.StaticTokenSource(req.Token)
	opt := []option.ClientOption{option.WithTokenSource(tokenSource)}
	srv, err := drive.NewService(ctx, opt...)

	return srv, err
}
