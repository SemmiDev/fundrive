package main

import (
    "context"
    "fmt"
    "github.com/semmidev/fundrive"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "log"
    "os"
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

    // example dummy data
    userID := "yyy"
    testFile, err := os.Open("testdata/author.txt")

    ctx := context.Background()

    // example upload
    file, err := fundriveService.UploadFile(ctx, &fundrive.UploadFileRequest{
        UserID:   userID,
        FileName: "Author",
        FileData: testFile,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("------------------------------------")
    fmt.Println(file)
    fmt.Printf("------------------------------------\n\n")

    // example get file
    savedFile, err := fundriveService.GetFileWithURL(ctx, &fundrive.GetFileRequest{
        UserID: userID,
        FileID: file.Id,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("------------------------------------")
    fmt.Println(savedFile)
    fmt.Printf("------------------------------------\n\n")

    // get storage info
    capacity, err := fundriveService.GetStorageInfo(nil, &fundrive.GetStorageInfoRequest{
        UserID: userID, // assume yyy is your valid user id in db ya >_
    })
    if err != nil {
        log.Fatal(err)
    }

    capacity.FormatTwoDigits()

    fmt.Println("------------------------------------")
    fmt.Println(capacity)
    fmt.Printf("------------------------------------\n\n")
}
