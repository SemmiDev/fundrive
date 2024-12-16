package fundrive

import (
    "fmt"
    "golang.org/x/oauth2"
    "time"
)

type OAuthToken struct {
    ID     string `json:"id" gorm:"column:id;type:char(26);primaryKey"`
    UserID string `json:"user_id" gorm:"column:user_id;type:char(255)"`
    Email  string `json:"email" gorm:"column:email;type:varchar(255)"`

    // provided by google
    AccessToken  string    `json:"access_token" gorm:"column:access_token;type:longtext"`
    RefreshToken string    `json:"refresh_token" gorm:"column:refresh_token;type:longtext"`
    TokenType    string    `json:"token_type" gorm:"column:token_type;type:varchar(50)"`
    Expiry       time.Time `json:"expiry" gorm:"column:expiry;type:timestamp"`
}

// TableName returns the table name
func (o *OAuthToken) TableName() string {
    return "fundrive_oauth_tokens"
}

// ToOAuth2Token converts OAuthToken to oauth2.Token
func (o *OAuthToken) ToOAuth2Token(encryption *TokenEncryption) (*oauth2.Token, error) {
    accessToken, err := encryption.Decrypt(o.AccessToken)
    if err != nil {
        return nil, fmt.Errorf("failed to decrypt access token: %w", err)
    }

    var refreshToken string
    if o.RefreshToken != "" {
        refreshToken, err = encryption.Decrypt(o.RefreshToken)
        if err != nil {
            return nil, fmt.Errorf("failed to decrypt refresh token: %w", err)
        }
    }

    return &oauth2.Token{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        TokenType:    o.TokenType,
        Expiry:       o.Expiry,
    }, nil
}

// FromOAuth2Token updates OAuthToken from oauth2.Token
func (o *OAuthToken) FromOAuth2Token(token *oauth2.Token, encryption *TokenEncryption) error {
    accessToken, err := encryption.Encrypt(token.AccessToken)
    if err != nil {
        return fmt.Errorf("failed to encrypt access token: %w", err)
    }

    var refreshToken string
    if token.RefreshToken != "" {
        refreshToken, err = encryption.Encrypt(token.RefreshToken)
        if err != nil {
            return fmt.Errorf("failed to encrypt refresh token: %w", err)
        }
    }

    o.AccessToken = accessToken
    o.RefreshToken = refreshToken
    o.TokenType = token.TokenType
    o.Expiry = token.Expiry

    return nil
}

func NewOauth2Token(accessToken string, refreshToken string, expiry time.Time) *oauth2.Token {
    return &oauth2.Token{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        TokenType:    "Bearer",
        Expiry:       expiry,
    }
}
