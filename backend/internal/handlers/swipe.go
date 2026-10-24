package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"petmatch.local/backend/internal/service"
)

type SwipeHandler struct {
	service *service.SwipeService
}

func NewSwipeHandler(s *service.SwipeService) *SwipeHandler { return &SwipeHandler{service: s} }

type SwipeInput struct {
	PetID  int  `json:"pet_id"`
	IsLike bool `json:"is_like"`
}

func (h *SwipeHandler) Swipe(c *gin.Context) {
	var in SwipeInput
	if !bindJSON(c, &in) {
		return
	}
	if in.PetID <= 0 {
		abort(c, http.StatusBadRequest, "invalid_input", "Укажите pet_id")
		return
	}
	res, err := h.service.Swipe(userID(c), in.PetID, in.IsLike)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

type OwnerSwipeInput struct {
	UserID int  `json:"user_id"`
	PetID  int  `json:"pet_id"`
	IsLike bool `json:"is_like"`
}

func (h *SwipeHandler) OwnerSwipe(c *gin.Context) {
	var in OwnerSwipeInput
	if !bindJSON(c, &in) {
		return
	}
	if in.PetID <= 0 || in.UserID <= 0 {
		abort(c, http.StatusBadRequest, "invalid_input", "Укажите user_id и pet_id")
		return
	}
	res, err := h.service.OwnerSwipe(userID(c), in.UserID, in.PetID, in.IsLike)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *SwipeHandler) Matches(c *gin.Context) {
	matches, err := h.service.GetMatches(userID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, matches)
}
