package main

import (
    "context"
    "fmt"
    "github.com/semmidev/fundrive"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "time"
)

func main() {

    // create a db connection
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", "root", "", "localhost", "3307", "fundrive")
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

    fundrive.PanicIfNeeded(err)

    // create a fundrive service
    fundriveService, err := fundrive.New(
        fundrive.WithDB(db),
        fundrive.WithServiceAccountFilePath("service-account.json"),
        fundrive.WithEncryptionKey("12345678901234567890123456789012"),
    )

    fundrive.PanicIfNeeded(err)

    err = fundriveService.OAuthService.SaveToken(context.Background(), &fundrive.SaveTokenRequest{
        UserID: "user-dev-2",
        Email:  "user-dev-2@gmail.com",
        Token:  fundrive.NewOauth2Token("access-token-example", "refresh-token-example", time.Now()),
    })

    fundrive.PanicIfNeeded(err)

    token, err := fundriveService.OAuthService.GetToken(context.Background(), &fundrive.GetTokenRequest{
        UserID: "user-dev-2",
        Email:  "user-dev-2@gmail.com",
    })

    fundrive.PanicIfNeeded(err)

    fmt.Println(token.AccessToken == "access-token-example")
    fmt.Println(token.RefreshToken == "refresh-token-example")

    err = fundriveService.OAuthService.SaveToken(context.Background(), &fundrive.SaveTokenRequest{
        UserID: "user-dev-2",
        Email:  "user-dev-3@gmail.com",
        Token:  fundrive.NewOauth2Token("access-token-example-2", "refresh-token-example-2", time.Now()),
    })

    fundrive.PanicIfNeeded(err)

    token, err = fundriveService.OAuthService.GetToken(context.Background(), &fundrive.GetTokenRequest{
        UserID: "user-dev-2",
        Email:  "user-dev-3@gmail.com",
    })

    fundrive.PanicIfNeeded(err)

    fmt.Println(token.AccessToken == "access-token-example-2")
    fmt.Println(token.RefreshToken == "refresh-token-example-2")
}
