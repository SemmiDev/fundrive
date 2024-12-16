package main

import (
	"fmt"
	"github.com/semmidev/fundrive"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
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

	// and done, you can use the services!
	_ = fundriveService
}
