package fundrive

import (
	"golang.org/x/oauth2"
	"time"
)

type OAuthToken struct {
	ID     string `json:"id" gorm:"column:id;type:char(26);primaryKey"`
	UserID string `json:"user_id" gorm:"column:user_id;type:char(255)"`
	Email  string `json:"email" gorm:"column:email;type:varchar(255)"`

	// provided by google
	AccessToken  string    `json:"access_token" gorm:"column:access_token;type:text"`
	TokenType    string    `json:"token_type" gorm:"column:token_type;type:varchar(50)"`
	RefreshToken string    `json:"refresh_token" gorm:"column:refresh_token;type:text"`
	Expiry       time.Time `json:"expiry" gorm:"column:expiry;type:timestamp"`
}

// TableName returns the table name
func (o *OAuthToken) TableName() string {
	return "fundrive_oauth_tokens"
}

// ToOAuth2Token converts OAuthToken to oauth2.Token
func (o *OAuthToken) ToOAuth2Token() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  o.AccessToken,
		TokenType:    o.TokenType,
		RefreshToken: o.RefreshToken,
		Expiry:       o.Expiry,
	}
}

// FromOAuth2Token updates OAuthToken from oauth2.Token
func (o *OAuthToken) FromOAuth2Token(token *oauth2.Token) {
	o.AccessToken = token.AccessToken
	o.TokenType = token.TokenType
	o.RefreshToken = token.RefreshToken
	o.Expiry = token.Expiry
}
