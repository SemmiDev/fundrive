package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/semmidev/fundrive"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
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
		fundrive.WithEncryptionKey("F7JI1y0TrkFnoeVMIONKIwAEshLrJqOy"),
	)

	fundrive.PanicIfNeeded(err)

	handler := fundrive.NewOAuthHandler(fundriveService.OauthConfig, fundriveService.DB, fundriveService.TokenEncryptor)
	app := fiber.New()
	app.Get("/auth/google/authorize", handler.AuthorizeHandler)
	app.Get("/auth/google/callback", handler.AuthorizeCallbackHandler)

	log.Println("Listening on port 3000")
	app.Listen(":3000")
}
