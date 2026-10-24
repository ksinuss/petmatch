// Package handlers — HTTP-слой (Gin): разбор запросов, вызов сервисов, формирование ответов.
package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"petmatch.local/backend/internal/service"
)

// ErrorBody — единый формат ошибки для всех эндпоинтов.
type ErrorBody struct {
	Error string `json:"error"` // сообщение для пользователя
	Code  string `json:"code"`  // машинный код: invalid_input, unauthorized, forbidden, not_found, conflict, internal
}

func abort(c *gin.Context, status int, code, msg string) {
	c.AbortWithStatusJSON(status, ErrorBody{Error: msg, Code: code})
}

// respondError превращает доменную ошибку сервиса в HTTP-статус.
// Неизвестные ошибки логируются, а пользователю отдаётся нейтральное сообщение.
func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		abort(c, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, service.ErrUnauthorized):
		abort(c, http.StatusUnauthorized, "unauthorized", err.Error())
	case errors.Is(err, service.ErrForbidden):
		abort(c, http.StatusForbidden, "forbidden", err.Error())
	case errors.Is(err, service.ErrNotFound):
		abort(c, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, service.ErrConflict):
		abort(c, http.StatusConflict, "conflict", err.Error())
	default:
		log.Printf("internal error: %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
		abort(c, http.StatusInternalServerError, "internal", "Внутренняя ошибка сервера, попробуйте позже")
	}
}

// bindJSON разбирает тело запроса; при ошибке сам отвечает 400 и возвращает false.
func bindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		abort(c, http.StatusBadRequest, "invalid_input", "Некорректный JSON в теле запроса")
		return false
	}
	return true
}

// pathID читает числовой параметр пути; при ошибке отвечает 400.
func pathID(c *gin.Context, name string) (int, bool) {
	id, err := strconv.Atoi(c.Param(name))
	if err != nil || id <= 0 {
		abort(c, http.StatusBadRequest, "invalid_input", "Некорректный идентификатор в пути запроса")
		return 0, false
	}
	return id, true
}
