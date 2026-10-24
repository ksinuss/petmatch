// Package app собирает приложение: репозитории → сервисы → обработчики → маршруты.
package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"petmatch.local/backend/internal/config"
	"petmatch.local/backend/internal/handlers"
	"petmatch.local/backend/internal/openapi"
	"petmatch.local/backend/internal/repository"
	"petmatch.local/backend/internal/seed"
	"petmatch.local/backend/internal/service"
	"petmatch.local/backend/internal/storage"
)

// New создаёт HTTP-обработчик всего API.
func New(cfg config.Config, db *sql.DB) (http.Handler, error) {
	// --- хранилище фото: S3, если заданы ключи, иначе локальный диск ---
	var photoStorage service.PhotoStorage
	if cfg.UseS3() {
		photoStorage = storage.NewS3(cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket)
		log.Println("photo storage: Yandex Object Storage, bucket", cfg.S3Bucket)
	} else {
		local, err := storage.NewLocal(cfg.UploadDir)
		if err != nil {
			return nil, err
		}
		photoStorage = local
		log.Println("photo storage: local dir", cfg.UploadDir)
	}

	// --- репозитории ---
	accountRepo := repository.NewAccountRepository(db)
	petRepo := repository.NewPetRepository(db)
	swipeRepo := repository.NewSwipeRepository(db)
	prefRepo := repository.NewPreferencesRepository(db)

	// --- сервисы (бизнес-логика) ---
	tokens := service.NewTokenManager(cfg.AuthSecret, time.Duration(cfg.TokenTTLHrs)*time.Hour)
	authSvc := service.NewAuthService(accountRepo, tokens)
	accountSvc := service.NewAccountService(accountRepo)
	petSvc := service.NewPetService(petRepo)
	photoSvc := service.NewPhotoService(petRepo, photoStorage, petSvc)
	swipeSvc := service.NewSwipeService(swipeRepo)
	prefSvc := service.NewPreferencesService(prefRepo)

	if cfg.SeedDemo {
		if err := seed.Run(accountRepo, petRepo, authSvc, petSvc); err != nil {
			return nil, fmt.Errorf("seed demo data: %w", err)
		}
	}

	// --- HTTP-слой ---
	authH := handlers.NewAuthHandler(authSvc)
	accountH := handlers.NewAccountHandler(accountSvc)
	petH := handlers.NewPetHandler(petSvc, photoSvc)
	swipeH := handlers.NewSwipeHandler(swipeSvc)
	prefH := handlers.NewPreferencesHandler(prefSvc)

	r := gin.New()
	r.Use(gin.Logger(), gin.CustomRecovery(func(c *gin.Context, _ any) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, handlers.ErrorBody{
			Error: "Внутренняя ошибка сервера, попробуйте позже", Code: "internal",
		})
	}))
	r.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		MaxAge:       12 * time.Hour,
	}))
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, handlers.ErrorBody{Error: "Маршрут не найден", Code: "not_found"})
	})

	// Служебные эндпоинты.
	r.GET("/health", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db_unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "demo": cfg.SeedDemo})
	})
	r.GET("/openapi.yaml", func(c *gin.Context) { c.Data(http.StatusOK, "application/yaml; charset=utf-8", openapi.Spec) })
	r.GET("/docs", func(c *gin.Context) { c.Data(http.StatusOK, "text/html; charset=utf-8", openapi.SwaggerUI) })
	r.StaticFS("/uploads", http.Dir(cfg.UploadDir))

	// Публичные маршруты.
	r.POST("/auth/register", authH.Register)
	r.POST("/auth/login", authH.Login)

	// Маршруты, требующие токен.
	api := r.Group("/", handlers.AuthRequired(tokens))
	api.GET("/me", accountH.GetMe)
	api.PUT("/me", accountH.UpdateMe)
	api.GET("/users/:id", accountH.GetUser)
	api.GET("/users/:id/pets", petH.ByOwner)

	api.POST("/pets", petH.Create)
	api.GET("/pets/feed", petH.Feed)
	api.GET("/pets/:id", petH.Get)
	api.POST("/pets/:id/photos", petH.UploadPhoto)
	api.GET("/my-pets", petH.Mine)

	api.POST("/swipes", swipeH.Swipe)
	api.POST("/owner-swipes", swipeH.OwnerSwipe)
	api.GET("/matches", swipeH.Matches)

	api.GET("/preferences", prefH.Get)
	api.POST("/preferences", prefH.Set)

	return r, nil
}
