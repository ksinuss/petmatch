package app_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"petmatch.local/backend/internal/app"
	"petmatch.local/backend/internal/config"
	"petmatch.local/backend/internal/database"
)

// client — тонкая обёртка над httptest для чтения сценариев «как у пользователя».
type client struct {
	t       *testing.T
	handler http.Handler
	token   string
}

func (c *client) do(method, path string, body any) (int, map[string]any, []any) {
	c.t.Helper()
	var rdr io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rdr)
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	return c.serve(req)
}

func (c *client) serve(req *http.Request) (int, map[string]any, []any) {
	c.t.Helper()
	w := httptest.NewRecorder()
	c.handler.ServeHTTP(w, req)

	raw := w.Body.Bytes()
	var obj map[string]any
	var arr []any
	if len(raw) > 0 && raw[0] == '[' {
		_ = json.Unmarshal(raw, &arr)
	} else {
		_ = json.Unmarshal(raw, &obj)
	}
	return w.Code, obj, arr
}

func newServer(t *testing.T) http.Handler {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		AuthSecret: "test-secret", TokenTTLHrs: 1,
		UploadDir: filepath.Join(t.TempDir(), "uploads"), CORSOrigins: []string{"http://localhost"},
	}
	h, err := app.New(cfg, db)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func newClient(t *testing.T, h http.Handler) *client { return &client{t: t, handler: h} }

func (c *client) register(email, name, role string) int {
	c.t.Helper()
	code, body, _ := c.do("POST", "/auth/register", map[string]any{
		"email": email, "password": "secret1", "display_name": name, "role": role,
	})
	if code != http.StatusCreated {
		c.t.Fatalf("register %s: %d %v", email, code, body)
	}
	c.token = body["token"].(string)
	return int(body["user_id"].(float64))
}

func num(m map[string]any, key string) int { return int(m[key].(float64)) }

// pngFile — минимальный PNG, достаточный для определения типа по сигнатуре.
var pngFile = append([]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, make([]byte, 64)...)

func (c *client) uploadPhoto(petID int, data []byte) (int, map[string]any) {
	c.t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("photo", "pet.png")
	_, _ = fw.Write(data)
	mw.Close()
	req := httptest.NewRequest("POST", fmt.Sprintf("/pets/%d/photos", petID), &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.token)
	code, obj, _ := c.serve(req)
	return code, obj
}

// TestEndToEndScenario — главный сценарий курсовой:
// хозяин публикует питомца → пользователь находит его в ленте, лайкает → мэтч.
func TestEndToEndScenario(t *testing.T) {
	h := newServer(t)

	// 1. Приют регистрируется и создаёт анкету с фото.
	owner := newClient(t, h)
	ownerID := owner.register("owner@test.ru", "Мария", "user")

	code, pet, _ := owner.do("POST", "/pets", map[string]any{
		"name": "Рич", "type": "dog", "breed": "Ретривер", "age": 1, "description": "Добрый щенок",
	})
	if code != http.StatusCreated || pet["adoption_mode"] != "open" {
		t.Fatalf("создание питомца: %d %v", code, pet)
	}
	petID := num(pet, "id")

	if code, resp := owner.uploadPhoto(petID, pngFile); code != http.StatusCreated || !strings.HasPrefix(resp["url"].(string), "/uploads/") {
		t.Fatalf("загрузка фото: %d %v", code, resp)
	}

	// Свою анкету владелец в ленте не видит.
	if _, _, feed := owner.do("GET", "/pets/feed", nil); len(feed) != 0 {
		t.Fatalf("владелец не должен видеть своих питомцев в ленте: %v", feed)
	}

	// 2. Пользователь регистрируется и видит питомца в ленте вместе с фото и именем владельца.
	user := newClient(t, h)
	userID := user.register("anna@test.ru", "Анна", "user")

	_, _, feed := user.do("GET", "/pets/feed", nil)
	if len(feed) != 1 {
		t.Fatalf("в ленте должен быть 1 питомец, получили %d", len(feed))
	}
	card := feed[0].(map[string]any)
	if card["name"] != "Рич" || card["owner_name"] != "Мария" || len(card["photos"].([]any)) != 1 {
		t.Fatalf("неполная карточка: %v", card)
	}

	// 3. Лайк в режиме open сразу даёт мэтч.
	code, res, _ := user.do("POST", "/swipes", map[string]any{"pet_id": petID, "is_like": true})
	if code != http.StatusOK || res["match"] != true {
		t.Fatalf("ожидали мэтч: %d %v", code, res)
	}
	if num(res, "match_id") == 0 {
		t.Fatal("в ответе должен быть match_id")
	}

	// Питомец исчез из ленты.
	if _, _, feed := user.do("GET", "/pets/feed", nil); len(feed) != 0 {
		t.Fatalf("оценённый питомец не должен возвращаться в ленту: %v", feed)
	}

	// 4. Мэтч виден обеим сторонам, у каждого — со своим собеседником.
	_, _, userMatches := user.do("GET", "/matches", nil)
	_, _, ownerMatches := owner.do("GET", "/matches", nil)
	if len(userMatches) != 1 || len(ownerMatches) != 1 {
		t.Fatalf("мэтч должен быть у обоих: %d / %d", len(userMatches), len(ownerMatches))
	}
	if p := userMatches[0].(map[string]any)["partner"].(map[string]any); num(p, "id") != ownerID {
		t.Errorf("у пользователя собеседник — хозяин, получили %v", p)
	}
	if p := ownerMatches[0].(map[string]any)["partner"].(map[string]any); num(p, "id") != userID {
		t.Errorf("у хозяина собеседник — пользователь, получили %v", p)
	}

}

func TestErrorScenarios(t *testing.T) {
	h := newServer(t)
	anon := newClient(t, h)

	t.Run("защищённые маршруты требуют токен", func(t *testing.T) {
		for _, path := range []string{"/me", "/pets/feed", "/matches", "/my-pets"} {
			if code, body, _ := anon.do("GET", path, nil); code != 401 || body["code"] != "unauthorized" {
				t.Errorf("%s без токена: %d %v", path, code, body)
			}
		}
	})

	t.Run("подделанный токен отклоняется", func(t *testing.T) {
		fake := &client{t: t, handler: h, token: "MQ.fake-signature"}
		if code, _, _ := fake.do("GET", "/me", nil); code != 401 {
			t.Errorf("ожидали 401, получили %d", code)
		}
	})

	t.Run("старый заголовок X-User-ID больше не работает", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/me", nil)
		req.Header.Set("X-User-ID", "1")
		if code, _, _ := anon.serve(req); code != 401 {
			t.Errorf("ожидали 401, получили %d", code)
		}
	})

	t.Run("регистрация: некорректные данные", func(t *testing.T) {
		bad := []map[string]any{
			{"email": "oops", "password": "secret1", "display_name": "A", "role": "user"},
			{"email": "a@b.ru", "password": "1", "display_name": "A", "role": "user"},
			{"email": "a@b.ru", "password": "secret1", "display_name": "A", "role": "hacker"},
			{"email": "a@b.ru", "password": "secret1", "display_name": "", "role": "user"},
		}
		for _, in := range bad {
			if code, body, _ := anon.do("POST", "/auth/register", in); code != 400 || body["error"] == "" {
				t.Errorf("%v: %d %v", in, code, body)
			}
		}
		req := httptest.NewRequest("POST", "/auth/register", strings.NewReader("{не json"))
		if code, body, _ := anon.serve(req); code != 400 || body["code"] != "invalid_input" {
			t.Errorf("битый JSON: %d %v", code, body)
		}
	})

	t.Run("дубликат email и неверный пароль", func(t *testing.T) {
		u := newClient(t, h)
		u.register("dup@test.ru", "Дуб", "user")
		if code, body, _ := anon.do("POST", "/auth/register", map[string]any{
			"email": "DUP@test.ru", "password": "secret1", "display_name": "X", "role": "user",
		}); code != 409 || body["code"] != "conflict" {
			t.Errorf("дубликат: %d %v", code, body)
		}
		if code, _, _ := anon.do("POST", "/auth/login", map[string]any{"email": "dup@test.ru", "password": "wrong-pass"}); code != 401 {
			t.Errorf("неверный пароль: %d", code)
		}
		if code, body, _ := anon.do("POST", "/auth/login", map[string]any{"email": "dup@test.ru", "password": "secret1"}); code != 200 || body["token"] == "" {
			t.Errorf("вход: %d %v", code, body)
		}
	})

	t.Run("питомец: валидация, чужие права, повторный свайп", func(t *testing.T) {
		owner := newClient(t, h)
		owner.register("owner@test.ru", "Владелец", "user")

		if code, _, _ := owner.do("POST", "/pets", map[string]any{"name": "", "type": "dog"}); code != 400 {
			t.Errorf("пустая кличка: %d", code)
		}
		if code, _, _ := owner.do("POST", "/pets", map[string]any{"name": "Вася", "type": "dragon"}); code != 400 {
			t.Errorf("неверный тип: %d", code)
		}
		_, pet, _ := owner.do("POST", "/pets", map[string]any{"name": "Вася", "type": "cat"})
		petID := num(pet, "id")

		// владелец не может свайпать своего питомца
		if code, _, _ := owner.do("POST", "/swipes", map[string]any{"pet_id": petID, "is_like": true}); code != 403 {
			t.Errorf("свой питомец: %d", code)
		}
		// несуществующий питомец
		if code, _, _ := owner.do("GET", "/pets/9999", nil); code != 404 {
			t.Errorf("нет питомца: %d", code)
		}
		if code, _, _ := owner.do("GET", "/pets/abc", nil); code != 400 {
			t.Errorf("некорректный id: %d", code)
		}

		other := newClient(t, h)
		other.register("other@test.ru", "Чужой", "user")
		// чужой не может грузить фото
		if code, _ := other.uploadPhoto(petID, pngFile); code != 403 {
			t.Errorf("чужое фото: %d", code)
		}
		// не-картинка отклоняется
		if code, _ := owner.uploadPhoto(petID, []byte("<script>alert(1)</script>")); code != 400 {
			t.Errorf("не картинка: %d", code)
		}

		if code, _, _ := other.do("POST", "/swipes", map[string]any{"pet_id": petID, "is_like": false}); code != 200 {
			t.Errorf("первый свайп: %d", code)
		}
		if code, _, _ := other.do("POST", "/swipes", map[string]any{"pet_id": petID, "is_like": true}); code != 409 {
			t.Errorf("повторный свайп: %d", code)
		}
	})

	t.Run("публичный профиль не содержит email", func(t *testing.T) {
		u := newClient(t, h)
		id := u.register("secret@mail.ru", "Скрытный", "user")
		code, body, _ := u.do("GET", fmt.Sprintf("/users/%d", id), nil)
		if code != 200 {
			t.Fatalf("%d", code)
		}
		if _, leaked := body["email"]; leaked {
			t.Error("email не должен попадать в публичный профиль")
		}
	})

	t.Run("служебные эндпоинты и неизвестный маршрут", func(t *testing.T) {
		if code, body, _ := anon.do("GET", "/health", nil); code != 200 || body["status"] != "ok" {
			t.Errorf("health: %d %v", code, body)
		}
		if code, _, _ := anon.do("GET", "/no-such-route", nil); code != 404 {
			t.Errorf("404: %d", code)
		}
		req := httptest.NewRequest("GET", "/openapi.yaml", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != 200 || !strings.Contains(w.Body.String(), "openapi: 3.0.3") {
			t.Errorf("openapi.yaml: %d", w.Code)
		}
	})
}

func TestStrictModeFlow(t *testing.T) {
	h := newServer(t)
	owner, user := newClient(t, h), newClient(t, h)
	ownerID := owner.register("o@strict.ru", "Приют", "user")
	_ = ownerID
	userID := user.register("u@strict.ru", "Пётр", "user")

	_, pet, _ := owner.do("POST", "/pets", map[string]any{"name": "Барон", "type": "dog", "adoption_mode": "strict"})
	petID := num(pet, "id")

	_, res, _ := user.do("POST", "/swipes", map[string]any{"pet_id": petID, "is_like": true})
	if res["match"] == true {
		t.Fatal("в strict-режиме лайк пользователя не должен сразу давать мэтч")
	}
	_, res, _ = owner.do("POST", "/owner-swipes", map[string]any{"user_id": userID, "pet_id": petID, "is_like": true})
	if res["match"] != true {
		t.Fatalf("после лайка владельца должен появиться мэтч: %v", res)
	}
}
