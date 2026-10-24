package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"petmatch.local/backend/internal/service"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(s *service.AuthService) *AuthHandler { return &AuthHandler{service: s} }

type RegisterInput struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func authResponse(res service.AuthResult) gin.H {
	return gin.H{
		"token":   res.Token,
		"user_id": res.Account.ID,
		"user":    res.Account,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var in RegisterInput
	if !bindJSON(c, &in) {
		return
	}
	res, err := h.service.Register(in.Email, in.Password, in.Role, in.DisplayName)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, authResponse(res))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var in LoginInput
	if !bindJSON(c, &in) {
		return
	}
	res, err := h.service.Login(in.Email, in.Password)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, authResponse(res))
}
