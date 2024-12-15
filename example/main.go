package main

import (
    "fmt"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/template/html/v2"
    "github.com/semmidev/fundrive"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "log"
    "net/http"
)

func main() {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", "root", "", "localhost", "3307", "fundrive")
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }

    fundriveService, err := fundrive.New(
        fundrive.WithDB(db),
        fundrive.WithServiceAccountFilePath("service-account.json"),
    )

    if err != nil {
        log.Fatal(err)
    }

    fundriveHandler := fundrive.NewOAuthHandler(fundriveService.OauthConfig, db)

    engine := html.New("./views", ".html")
    app := fiber.New(fiber.Config{
        Views: engine,
    })

    app.Get("/", func(c *fiber.Ctx) error {
        return c.Render("index", fiber.Map{
            "Title": "Welcome, Please Login!",
        })
    })

    app.Get("/auth/google/authorize", fundriveHandler.AuthorizeHandler)
    app.Get("/auth/google/callback", fundriveHandler.AuthorizeCallbackHandler)

    app.Get("/auth/google/logout", func(c *fiber.Ctx) error {
        return c.Redirect("/auth/google/authorize", http.StatusFound)
    })

    app.Get("/home", func(c *fiber.Ctx) error {
        oauthTokens := make([]fundrive.OAuthToken, 0)
        if err := db.Find(&oauthTokens).Error; err != nil {
            return err
        }

        listStorageInfo := make([]fundrive.StorageInfo, 0)
        for _, token := range oauthTokens {
            storageInfo, err := fundriveService.GetStorageInfo(c.Context(), &fundrive.GetStorageInfoRequest{
                UserID: token.UserID,
                Email:  token.Email,
            })

            if err != nil {
                return err
            }

            listStorageInfo = append(listStorageInfo, *storageInfo)
        }

        return c.Render("home", fiber.Map{
            "Title":           "Welcome, Please Login!",
            "ListStorageInfo": listStorageInfo,
        })
    })

    log.Println("Server started on http://localhost:3000")
    if err := app.Listen(":3000"); err != nil {
        log.Fatal(err)
    }
}
