package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"petmatch.local/backend/internal/service"
)

const userIDKey = "userID"

// AuthRequired проверяет заголовок Authorization: Bearer <token>
// и кладёт id пользователя в контекст запроса.
func AuthRequired(tokens *service.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			abort(c, http.StatusUnauthorized, "unauthorized", "Требуется авторизация")
			return
		}
		id, err := tokens.Parse(strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			respondError(c, err)
			return
		}
		c.Set(userIDKey, id)
		c.Next()
	}
}

// userID возвращает id авторизованного пользователя (после AuthRequired).
func userID(c *gin.Context) int {
	return c.GetInt(userIDKey)
}
