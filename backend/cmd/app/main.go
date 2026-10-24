// PetMatch — сервис знакомства людей с питомцами из приютов (свайпы, мэтчи, чат).
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"petmatch.local/backend/internal/app"
	"petmatch.local/backend/internal/config"
	"petmatch.local/backend/internal/database"
)

func main() {
	cfg := config.Load()
	if cfg.AuthSecret == config.DefaultAuthSecret {
		log.Println("WARNING: AUTH_SECRET не задан, используется небезопасное значение для разработки")
	}

	db, err := database.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Fatal(err)
	}

	handler, err := app.New(cfg, db)
	if err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Println("PetMatch API слушает :" + cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	// Плавная остановка по SIGTERM (docker stop) и Ctrl+C.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("shutdown:", err)
	}
}
