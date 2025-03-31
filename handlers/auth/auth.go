package auth

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func SignInHandler(pass string) gin.HandlerFunc {
	return func(c *gin.Context) {
		type AuthRequest struct {
			Password string `json:"password"`
		}

		var req AuthRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
			return
		}

		// Если пароль в окружении не установлен, пропускаем аутентификацию
		if pass == "" {
			c.JSON(http.StatusOK, gin.H{"token": "notoken"})
			return
		}

		if req.Password != pass {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный пароль"})
			return
		}

		// Генерация JWT
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"hash": fmt.Sprintf("%x", sha256.Sum256([]byte(pass))), // Хэш пароля
			"exp":  time.Now().Add(8 * time.Hour).Unix(),
		})

		tokenString, err := token.SignedString([]byte(pass))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации токена"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	}
}
func AuthMiddleware(pass string) gin.HandlerFunc {
	return func(c *gin.Context) {
		
		// Если пароль не установлен, пропускаем проверку
		if pass == "" {
			c.Next()
			return
		}

		tokenString, err := c.Cookie("token")
		if err != nil || tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
			return
		}

		// Парсим токен
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Неожиданный метод подписи")
			}
			return []byte(pass), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен"})
			return
		}

		// Проверяем хэш пароля
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			expectedHash := fmt.Sprintf("%x", sha256.Sum256([]byte(pass)))
			if claims["hash"] != expectedHash {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Токен устарел"})
				return
			}
		}

		c.Next()
	}
}
