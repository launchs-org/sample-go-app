package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const uploadsDir = "/uploads"

func main() {
	// /uploads ディレクトリが存在しない場合は作成
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Fatalf("uploads ディレクトリの作成に失敗: %v", err)
	}

	app := echo.New()
	app.Use(middleware.Logger())

	// 静的ファイルの配信
	app.Static("/", "statics")
	app.Static("/uploads", uploadsDir)

	// 画像一覧の取得
	app.GET("/files", func(c echo.Context) error {
		entries, err := os.ReadDir(uploadsDir)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, []string{})
		}
		var fileNames []string
		for _, entry := range entries {
			if !entry.IsDir() {
				fileNames = append(fileNames, entry.Name())
			}
		}
		return c.JSON(http.StatusOK, fileNames)
	})

	// ファイル削除
	app.DELETE("/files/:name", func(c echo.Context) error {
		name := filepath.Base(c.Param("name")) // パストラバーサル対策
		target := filepath.Join(uploadsDir, name)
		if err := os.Remove(target); err != nil {
			return c.String(http.StatusInternalServerError, "削除失敗")
		}
		return c.String(http.StatusOK, "削除成功")
	})

	// ファイルアップロードの処理
	app.POST("/upload", func(c echo.Context) error {
		file, err := c.FormFile("file")
		if err != nil {
			log.Println(err)
			return err
		}

		src, err := file.Open()
		if err != nil {
			log.Println(err)
			return err
		}
		defer src.Close()

		// MIMEタイプチェック（先頭512バイトで判定）
		buffer := make([]byte, 512)
		if _, err := src.Read(buffer); err != nil {
			return c.String(http.StatusInternalServerError, "ファイル読み込み失敗")
		}
		contentType := http.DetectContentType(buffer)
		if len(contentType) < 5 || contentType[:5] != "image" {
			return c.String(http.StatusBadRequest, "画像のみ許可されています")
		}
		if _, err := src.Seek(0, io.SeekStart); err != nil {
			return c.String(http.StatusInternalServerError, "シーク失敗")
		}

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
			return c.String(http.StatusBadRequest, "サポートされていない画像形式です")
		}

		// 絶対パスで保存
		fileName := uuid.New().String() + extension
		destPath := filepath.Join(uploadsDir, fileName)
		dst, err := os.Create(destPath)
		if err != nil {
			log.Println(err)
			return c.String(http.StatusInternalServerError, "ファイル作成失敗")
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			log.Println(err)
			return c.String(http.StatusInternalServerError, "保存失敗")
		}

		return c.String(http.StatusOK, "アップロード成功: "+fileName)
	})

	app.Logger.Fatal(app.Start(":1323"))
}