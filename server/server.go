package server

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/Jtrx1/go_final_project/handlers"
	"github.com/gin-gonic/gin"
)

// SetupRouter создает и настраивает роутер Gin
func SetupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	// 1. Сначала регистрируем API-маршруты. Особенности Gin
	r.GET("/api/nextdate", handlers.NextDateHandler)
	r.POST("/api/task", handlers.AddTask(db))
	// 2. Настраиваем статику через NoRoute. Особенности Gin
	r.NoRoute(func(c *gin.Context) {
		// Путь к запрашиваемому файлу
		filePath := filepath.Join("./web", c.Request.URL.Path)

		// Проверяем существование файла, если файл существует - отдаём его
		if _, err := os.Stat(filePath); err == nil {
			c.File(filePath)
			return
		}
	})

	return r
}
