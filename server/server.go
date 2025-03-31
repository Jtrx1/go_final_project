package server

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/Jtrx1/go_final_project/config"
	"github.com/Jtrx1/go_final_project/handlers"
	"github.com/Jtrx1/go_final_project/handlers/auth"
	"github.com/gin-gonic/gin"
)

// SetupRouter создает и настраивает роутер Gin
func SetupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()
	// 1. Сначала регистрируем API-маршруты. Особенности Gin
	r.POST("/api/signin", auth.SignInHandler(config.СheckEnv().Password))
	// Protected routes
	authGroup := r.Group("/api")
	authGroup.Use(auth.AuthMiddleware())
	{
		r.GET("/api/nextdate", handlers.NextDateHandler)
		r.GET("/api/tasks", handlers.GetTasks(db))
		r.POST("/api/task", handlers.AddTask(db))
		r.POST("/api/task/done", handlers.TaskDone(db))
		r.PUT("/api/task", handlers.EditTask(db))
		r.DELETE("/api/task", handlers.DeleteTask(db))
		r.GET("/api/task", handlers.GetTask(db))
	}
	// 2. Настраиваем статику через NoRoute. Особенности Gin
	r.NoRoute(func(c *gin.Context) {
		filePath := filepath.Join("./web", c.Request.URL.Path)
		if _, err := os.Stat(filePath); err == nil {
			c.File(filePath)
			return
		}
	})
	return r
}
