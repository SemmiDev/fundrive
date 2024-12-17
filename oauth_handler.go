package fundrive

import (
	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
	"net/http"
	"net/url"
)

type OAuthHandler struct {
	oauth2Config   *oauth2.Config
	db             *gorm.DB
	tokenEncryptor *TokenEncryption
}

func NewOAuthHandler(
	oauth2Config *oauth2.Config,
	db *gorm.DB,
	tokenEncryptor *TokenEncryption,
) OAuthHandler {
	return OAuthHandler{
		oauth2Config:   oauth2Config,
		db:             db,
		tokenEncryptor: tokenEncryptor,
	}
}

func (handler *OAuthHandler) Route(app *fiber.App) {
	app.Get("/auth/google/authorize", handler.AuthorizeHandler)
	app.Get("/auth/google/callback", handler.AuthorizeCallbackHandler)
}

type AuthorizeResponse struct {
	URL string `json:"url"`
}

func (handler *OAuthHandler) AuthorizeHandler(c *fiber.Ctx) error {
	redirectURL := c.Query("redirect_url")

	queryParams := url.Values{
		"redirect_url": {redirectURL},
	}

	// https://medium.com/starthinker/google-oauth-2-0-access-token-and-refresh-token-explained-cccf2fc0a6d9
	authCodeOpt := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline, // obtain the refresh token
		oauth2.ApprovalForce,     // forces the users to view the consent dialog
	}

	loginUrl := handler.oauth2Config.AuthCodeURL(queryParams.Encode(), authCodeOpt...)
	return c.Redirect(loginUrl, http.StatusFound)
}

func (handler *OAuthHandler) AuthorizeCallbackHandler(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")

	queryParams, _ := url.ParseQuery(state)
	redirectURL := queryParams.Get("redirect_url")

	token, err := handler.oauth2Config.Exchange(c.Context(), code)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"success": false,
			"message": http.StatusText(http.StatusInternalServerError),
			"details": err.Error(),
		})
	}

	oAuthConfig := OAuthConfig{
		DB:             handler.db,
		OAuth2Config:   handler.oauth2Config,
		TokenEncryptor: handler.tokenEncryptor,
	}

	oauthService, err := NewOAuthService(&oAuthConfig)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"success": false,
			"message": http.StatusText(http.StatusInternalServerError),
			"details": err.Error(),
		})
	}

	userInfo, err := oauthService.GetGoogleUserInfo(c.Context(), &GetUserInfoRequest{Token: token})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"success": false,
			"message": http.StatusText(http.StatusInternalServerError),
			"details": err.Error(),
		})
	}

	userInfo.ID = "user-dev" // dummy user, simulate multiple email with same ID

	saveTokenReq := SaveTokenRequest{
		UserID: userInfo.ID,
		Email:  userInfo.Email,
		Token:  token,
	}

	err = oauthService.SaveToken(c.Context(), &saveTokenReq)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"code":    http.StatusInternalServerError,
			"success": false,
			"message": http.StatusText(http.StatusInternalServerError),
			"details": err.Error(),
		})
	}

	if redirectURL != "" {
		return c.Redirect(redirectURL, http.StatusFound)
	}

	return c.Status(200).JSON(fiber.Map{
		"code":    http.StatusOK,
		"success": true,
		"message": http.StatusText(http.StatusOK),
		"details": userInfo,
	})
}
