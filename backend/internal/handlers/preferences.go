package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"petmatch.local/backend/internal/model"
	"petmatch.local/backend/internal/service"
)

type PreferencesHandler struct {
	service *service.PreferencesService
}

func NewPreferencesHandler(s *service.PreferencesService) *PreferencesHandler {
	return &PreferencesHandler{service: s}
}

func (h *PreferencesHandler) Set(c *gin.Context) {
	var in model.Preferences
	if !bindJSON(c, &in) {
		return
	}
	p, err := h.service.Set(userID(c), in)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *PreferencesHandler) Get(c *gin.Context) {
	p, err := h.service.Get(userID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}
