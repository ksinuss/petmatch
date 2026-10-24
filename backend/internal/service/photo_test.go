package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"petmatch.local/backend/internal/model"
)

type fakeStorage struct{ saved int }

func (f *fakeStorage) Save(_ context.Context, r io.Reader, ext, _ string) (string, error) {
	f.saved++
	return "/uploads/x" + ext, nil
}

type fakePhotoRepo struct{ urls []string }

func (f *fakePhotoRepo) AddPetPhoto(_ int, url string) error {
	f.urls = append(f.urls, url)
	return nil
}

// минимальный валидный PNG-заголовок: DetectContentType смотрит на сигнатуру
var pngBytes = append([]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, make([]byte, 32)...)

func newPhotoSvc() (*PhotoService, *fakeStorage, *fakePhotoRepo) {
	pets := NewPetService(&fakePetRepo{pets: map[int]model.Pet{1: {ID: 1, OwnerID: 5, Name: "Рич"}}})
	st, repo := &fakeStorage{}, &fakePhotoRepo{}
	return NewPhotoService(repo, st, pets), st, repo
}

func TestUploadPhotoOK(t *testing.T) {
	s, st, repo := newPhotoSvc()
	url, err := s.UploadPetPhoto(context.Background(), 5, 1, bytes.NewReader(pngBytes))
	if err != nil || url != "/uploads/x.png" || st.saved != 1 || len(repo.urls) != 1 {
		t.Fatalf("url=%q err=%v saved=%d", url, err, st.saved)
	}
}

func TestUploadPhotoRejects(t *testing.T) {
	s, st, _ := newPhotoSvc()
	ctx := context.Background()

	if _, err := s.UploadPetPhoto(ctx, 6, 1, bytes.NewReader(pngBytes)); !errors.Is(err, ErrForbidden) {
		t.Errorf("чужой питомец: %v", err)
	}
	if _, err := s.UploadPetPhoto(ctx, 5, 99, bytes.NewReader(pngBytes)); !errors.Is(err, ErrNotFound) {
		t.Errorf("нет питомца: %v", err)
	}
	if _, err := s.UploadPetPhoto(ctx, 5, 1, strings.NewReader("<html>не картинка</html>")); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("не картинка: %v", err)
	}
	if _, err := s.UploadPetPhoto(ctx, 5, 1, bytes.NewReader(nil)); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("пустой файл: %v", err)
	}
	big := append(append([]byte{}, pngBytes...), make([]byte, MaxPhotoSize)...)
	if _, err := s.UploadPetPhoto(ctx, 5, 1, bytes.NewReader(big)); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("слишком большой файл: %v", err)
	}
	if st.saved != 0 {
		t.Errorf("при ошибках в хранилище ничего не должно сохраняться, сохранено %d", st.saved)
	}
}
