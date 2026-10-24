package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

// PhotoStorage — куда физически кладутся файлы (локальный диск или S3).
type PhotoStorage interface {
	Save(ctx context.Context, data io.Reader, ext, contentType string) (url string, err error)
}

type PhotoRepository interface {
	AddPetPhoto(petID int, url string) error
}

// MaxPhotoSize — максимальный размер загружаемого фото.
const MaxPhotoSize = 5 << 20

var allowedImages = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type PhotoService struct {
	repo    PhotoRepository
	storage PhotoStorage
	pets    *PetService
}

func NewPhotoService(r PhotoRepository, st PhotoStorage, pets *PetService) *PhotoService {
	return &PhotoService{repo: r, storage: st, pets: pets}
}

// UploadPetPhoto проверяет права и тип файла, сохраняет его и привязывает к питомцу.
func (s *PhotoService) UploadPetPhoto(ctx context.Context, userID, petID int, file io.Reader) (string, error) {
	pet, err := s.pets.GetPet(petID)
	if err != nil {
		return "", err
	}
	if pet.OwnerID != userID {
		return "", forbidden("Добавлять фото может только владелец питомца")
	}

	data, err := io.ReadAll(io.LimitReader(file, MaxPhotoSize+1))
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", invalid("Файл пустой")
	}
	if len(data) > MaxPhotoSize {
		return "", invalid("Фото слишком большое (максимум 5 МБ)")
	}

	// Тип определяем по содержимому, а не по расширению или заголовку клиента.
	ctype := http.DetectContentType(data)
	ext, ok := allowedImages[ctype]
	if !ok {
		return "", invalid("Допустимы только изображения JPEG, PNG или WebP")
	}

	url, err := s.storage.Save(ctx, bytes.NewReader(data), ext, ctype)
	if err != nil {
		return "", err
	}
	if err := s.repo.AddPetPhoto(petID, url); err != nil {
		return "", err
	}
	return url, nil
}
