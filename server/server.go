package server

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// SetupRouter создает и настраивает роутер Gin
func SetupRouter() *gin.Engine {
	r := gin.Default()
	
	// Отдаем статические файлы из указанной папки
	r.StaticFS("/", http.Dir("./web"))
	
	return r
}