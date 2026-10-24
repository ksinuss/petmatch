package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"petmatch.local/backend/internal/service"
)

type AccountHandler struct {
	service *service.AccountService
}

func NewAccountHandler(s *service.AccountService) *AccountHandler { return &AccountHandler{service: s} }

func (h *AccountHandler) GetMe(c *gin.Context) {
	acc, err := h.service.GetMe(userID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, acc)
}

type UpdateMeInput struct {
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio"`
	Location    string `json:"location"`
}

func (h *AccountHandler) UpdateMe(c *gin.Context) {
	var in UpdateMeInput
	if !bindJSON(c, &in) {
		return
	}
	acc, err := h.service.UpdateMe(userID(c), in.DisplayName, in.Bio, in.Location)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, acc)
}

func (h *AccountHandler) GetUser(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	acc, err := h.service.GetPublic(id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, acc)
}
