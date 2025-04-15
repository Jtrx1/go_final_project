package server

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/Jtrx1/go_final_project/handlers"
	"github.com/Jtrx1/go_final_project/handlers/auth"
	"github.com/gin-gonic/gin"
)

// SetupRouter создает и настраивает роутер Gin
func SetupRouter(db *sql.DB, pass string) *gin.Engine {
	r := gin.Default()
	// Public routes
	r.POST("/api/signin", auth.SignInHandler(pass))
	r.GET("/api/nextdate", handlers.NextDateHandler)
	// Protected routes group
	authGroup := r.Group("/")
	authGroup.Use(auth.AuthMiddleware(pass))
	{
		authGroup.GET("/api/tasks", handlers.GetTasks(db))
		authGroup.POST("/api/task", handlers.AddTask(db))
		authGroup.POST("/api/task/done", handlers.TaskDone(db))
		authGroup.PUT("/api/task", handlers.EditTask(db))
		authGroup.DELETE("/api/task", handlers.DeleteTask(db))
		authGroup.GET("/api/task", handlers.GetTask(db))
	}

	// Static files
	r.NoRoute(func(c *gin.Context) {
		filePath := filepath.Join("./web", c.Request.URL.Path)
		if _, err := os.Stat(filePath); err == nil {
			c.File(filePath)
			return
		}
	})

	return r
}
