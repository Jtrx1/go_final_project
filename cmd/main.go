package main

import (
  "net/http"
	"github.com/gin-gonic/gin"
	
)

func main() {
	r := gin.Default()
    r.StaticFS("/", http.Dir("./web"))
  r.Run(":7154")
}
