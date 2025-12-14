package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/k-kanke/code-stash-server/internal/infra/router"
)

func main() {
	db, err := newDB()
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}
	defer db.Close()

	handler, err := InitializeHandler(db)
	if err != nil {
		log.Fatalf("failed to initialize handler: %v", err)
	}

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	router.RegisterRouter(e, handler)

	addr := getEnv("API_ADDR", ":8085")
	if err := e.Start(addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func newDB() (*sql.DB, error) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		host := getEnv("DB_HOST", "localhost")
		port := getEnv("DB_PORT", "5437")
		user := getEnv("DB_USER", "root")
		password := getEnv("DB_PASSWORD", "password")
		dbName := getEnv("DB_NAME", "codestash")
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbName)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
