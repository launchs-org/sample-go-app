package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	app := echo.New()
	app.Use(middleware.Logger())

	// 静的ファイルの配信
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", 0755)
	}
	app.Static("/", "statics")
	app.Static("/uploads", "uploads")

	// 画像一覧の取得
	app.GET("/files", func(context echo.Context) error {
		files, _ := os.ReadDir("uploads")
		var fileNames []string
		for _, file := range files {
			fileNames = append(fileNames, file.Name())
		}
		return context.JSON(http.StatusOK, fileNames)
	})

	// ファイル削除
	app.DELETE("/files/:name", func(context echo.Context) error {
		name := context.Param("name")
		err := os.Remove("uploads/" + name)
		if err != nil {
			return context.String(http.StatusInternalServerError, "削除失敗")
		}
		return context.String(http.StatusOK, "削除成功")
	})

	// ファイルアップロードの処理
	app.POST("/upload", func(context echo.Context) error {
		file, err := context.FormFile("file")
		if err != nil {
			log.Println(err)
			return err
		}
		
		// MIMEタイプチェック
		src, err := file.Open()
		if err != nil {
			log.Println(err)
			return err
		}
		defer src.Close()

		buffer := make([]byte, 512)
		_, _ = src.Read(buffer)
		contentType := http.DetectContentType(buffer)
		if contentType[:5] != "image" {
			return context.String(http.StatusBadRequest, "画像のみ許可されています")
		}
		_, _ = src.Seek(0, 0)

		// MIMEタイプから拡張子を決定
		var extension string
		switch contentType {
		case "image/jpeg":
			extension = ".jpg"
		case "image/png":
			extension = ".png"
		case "image/gif":
			extension = ".gif"
		default:
			return context.String(http.StatusBadRequest, "サポートされていない画像形式です")
		}

		// 保存
		fileName := uuid.New().String() + extension
		dst, err := os.Create("uploads/" + fileName)
		if err != nil {
			log.Println(err)
			return err
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			log.Println(err)
			return err
		}

		return context.String(http.StatusOK, "アップロード成功: "+fileName)
	})

	app.Logger.Fatal(app.Start(":1323"))
}
