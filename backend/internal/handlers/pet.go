package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"petmatch.local/backend/internal/model"
	"petmatch.local/backend/internal/service"
)

type PetHandler struct {
	service *service.PetService
	photos  *service.PhotoService
}

func NewPetHandler(s *service.PetService, p *service.PhotoService) *PetHandler {
	return &PetHandler{service: s, photos: p}
}

type CreatePetInput struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Breed        string `json:"breed"`
	Age          int    `json:"age"`
	Description  string `json:"description"`
	AdoptionMode string `json:"adoption_mode"`
}

func (h *PetHandler) Create(c *gin.Context) {
	var in CreatePetInput
	if !bindJSON(c, &in) {
		return
	}
	pet, err := h.service.CreatePet(model.Pet{
		OwnerID: userID(c), Name: in.Name, Type: in.Type, Breed: in.Breed,
		Age: in.Age, Description: in.Description, AdoptionMode: in.AdoptionMode,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, pet)
}

func (h *PetHandler) Get(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	pet, err := h.service.GetPet(id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, pet)
}

func (h *PetHandler) Mine(c *gin.Context) {
	pets, err := h.service.GetPetsByOwner(userID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, pets)
}

func (h *PetHandler) ByOwner(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	pets, err := h.service.GetPetsByOwner(id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, pets)
}

func (h *PetHandler) Feed(c *gin.Context) {
	pets, err := h.service.GetFeed(userID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, pets)
}

// UploadPhoto принимает multipart-форму с файлом в поле "photo".
func (h *PetHandler) UploadPhoto(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	// Жёсткий потолок на тело запроса, чтобы нельзя было залить гигабайты.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, service.MaxPhotoSize+1<<20)

	fh, err := c.FormFile("photo")
	if err != nil {
		abort(c, http.StatusBadRequest, "invalid_input", "Прикрепите файл в поле photo (не больше 5 МБ)")
		return
	}
	f, err := fh.Open()
	if err != nil {
		respondError(c, err)
		return
	}
	defer f.Close()

	url, err := h.photos.UploadPetPhoto(c.Request.Context(), userID(c), id, f)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"url": url})
}
