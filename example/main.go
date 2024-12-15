package main

import (
	"context"
	"fmt"
	"github.com/semmidev/fundrive"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

func main() {

	// create a db connection
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", "root", "", "localhost", "3307", "fundrive")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// create a fundrive service
	fundriveService, err := fundrive.New(
		fundrive.WithDB(db),
		fundrive.WithServiceAccountFilePath("service-account.json"),
	)

	if err != nil {
		log.Fatal(err)
	}

	// assume the client is already authenticated, and successfully retrieve a token, so lets save the token.
	oauthToken := fundrive.OAuthToken{
		Email:        "test2gmail.com",
		AccessToken:  "xxx",
		TokenType:    "Bearer",
		RefreshToken: "yyy",
		Expiry:       time.Now().Add(time.Hour * 24),
	}

	if err = fundriveService.OAuthService.SaveToken(context.Background(), &fundrive.SaveTokenRequest{
		UserID: "123",
		Email:  "xYV6u@example.com",
		Token:  oauthToken.ToOAuth2Token(),
	}); err != nil {
		log.Fatal(err)
	}

	// and done, you can use the services!
	_ = fundriveService
}
