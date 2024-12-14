package fundrive

import (
	"context"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func (service *GoogleDriveService) newDriveService(ctx context.Context, userID string) (*drive.Service, error) {
	token, err := service.oAuthService.GetToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	srv, err := service.newTokenService(ctx, userID, token)
	if err != nil {
		return nil, err
	}

	return srv, nil
}

// newTokenService creates a new Google Drive service using the provided token.
// If the token is invalid, it will refresh the token and create a new service.
func (service *GoogleDriveService) newTokenService(
	ctx context.Context,
	userID string,
	token *oauth2.Token,
) (*drive.Service, error) {
	if !token.Valid() {
		refreshedToken, err := service.oAuthService.RefreshToken(ctx, token)
		if err != nil {
			return nil, err
		}

		token = refreshedToken

		if err = service.oAuthService.SaveToken(ctx, userID, token); err != nil {
			return nil, err
		}
	}

	tokenSource := oauth2.StaticTokenSource(token)
	opt := []option.ClientOption{option.WithTokenSource(tokenSource)}
	srv, err := drive.NewService(ctx, opt...)

	return srv, err
}
