package fundrive

import (
    "github.com/gofiber/fiber/v2"
    "golang.org/x/oauth2"
    "net/http"
    "net/url"
)

type OAuthController struct {
    oauthService IOAuthService
    oauth2Config *oauth2.Config
}

func NewOAuthController(
    oauthService IOAuthService,
    oauth2Config *oauth2.Config,
) OAuthController {
    return OAuthController{
        oauthService: oauthService,
        oauth2Config: oauth2Config,
    }
}

func (controller *OAuthController) Route(app *fiber.App) {
    app.Get("/auth/google/authorize", controller.AuthorizeHandler)
    app.Get("/auth/google/callback", controller.AuthorizeCallbackHandler)
    app.Get("/auth/google/token", controller.GetTokenHandler)
    app.Get("/auth/google/token/exists", controller.TokenExistsHandler)
}

type AuthorizeResponse struct {
    URL string `json:"url"`
}

func (controller *OAuthController) AuthorizeHandler(c *fiber.Ctx) error {
    redirectURL := c.Query("redirect_url")
    userId := c.Query("user_id")
    if userId == "" {
        return c.Status(400).JSON(fiber.Map{
            "code":    http.StatusBadRequest,
            "success": false,
            "message": http.StatusText(http.StatusBadRequest),
            "details": "user_id is required",
        })
    }

    queryParams := url.Values{
        "user_id":      {userId},
        "redirect_url": {redirectURL},
    }

    // https://medium.com/starthinker/google-oauth-2-0-access-token-and-refresh-token-explained-cccf2fc0a6d9
    authCodeOpt := []oauth2.AuthCodeOption{
        oauth2.AccessTypeOffline, // obtain the refresh token
        oauth2.ApprovalForce,     // forces the users to view the consent dialog
    }

    loginUrl := controller.oauth2Config.AuthCodeURL(queryParams.Encode(), authCodeOpt...)

    return c.Status(200).JSON(WebResponse[AuthorizeResponse]{
        Code:    200,
        Status:  http.StatusText(200),
        Success: true,
        Data:    AuthorizeResponse{URL: loginUrl},
    })
}

func (controller *OAuthController) AuthorizeCallbackHandler(c *fiber.Ctx) error {
    code := c.Query("code")

    state := c.Query("state")
    queryParams, _ := url.ParseQuery(state)

    userId := queryParams.Get("user_id")
    redirectURL := queryParams.Get("redirect_url")

    token, err := controller.oauth2Config.Exchange(c.Context(), code)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "code":    http.StatusInternalServerError,
            "success": false,
            "message": http.StatusText(http.StatusInternalServerError),
            "details": err.Error(),
        })
    }

    err = controller.oauthService.SaveToken(c.Context(), userId, token)
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
        "code":    200,
        "success": true,
        "message": http.StatusText(200),
        "details": "Token has been saved",
    })
}

func (controller *OAuthController) GetTokenHandler(c *fiber.Ctx) error {
    userId := c.Query("user_id")
    if userId == "" {
        return c.Status(400).JSON(fiber.Map{
            "code":    http.StatusBadRequest,
            "success": false,
            "message": http.StatusText(http.StatusBadRequest),
            "details": "user_id is required",
        })
    }

    token, err := controller.oauthService.GetToken(c.Context(), userId)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "code":    http.StatusInternalServerError,
            "success": false,
            "message": http.StatusText(http.StatusInternalServerError),
            "details": err.Error(),
        })
    }

    return c.Status(200).JSON(fiber.Map{
        "code":    200,
        "success": true,
        "message": http.StatusText(200),
        "details": "Token has been retrieved",
        "data":    token,
    })
}

func (controller *OAuthController) TokenExistsHandler(c *fiber.Ctx) error {
    userId := c.Query("user_id")
    if userId == "" {
        return c.Status(400).JSON(fiber.Map{
            "code":    http.StatusBadRequest,
            "success": false,
            "message": http.StatusText(http.StatusBadRequest),
            "details": "user_id is required",
        })
    }

    exist, err := controller.oauthService.IsTokenExists(c.Context(), userId)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "code":    http.StatusInternalServerError,
            "success": false,
            "message": http.StatusText(http.StatusInternalServerError),
            "details": err.Error(),
        })
    }

    return c.Status(200).JSON(fiber.Map{
        "code":    200,
        "success": true,
        "message": http.StatusText(200),
        "data":    fiber.Map{"exists": exist},
    })
}

type WebResponse[T any] struct {
    Code    int    `json:"code"`
    Status  string `json:"message"`
    Success bool   `json:"success"`
    Data    T      `json:"data"`
}

type ErrorResponse struct {
    Code    int    `json:"code"`
    Success bool   `json:"success"`
    Message string `json:"message"`
    Details string `json:"details"`
}

func (e ErrorResponse) Error() string {
    return e.Message
}
