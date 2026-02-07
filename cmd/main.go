package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"test-psql/internal/app"
	"test-psql/internal/database"
	"test-psql/internal/http/handlers"
	"test-psql/internal/http/middleware"
	"test-psql/internal/queue"
	"test-psql/internal/repo"
	"test-psql/internal/service"
	"test-psql/pkg/env"
	"test-psql/pkg/logger"
)

func main() {
	// Инициализация логгера
	verbose := flag.Bool("v", false, "enable verbose logging")
	flag.Parse()

	if err := logger.Init(*verbose, "app.log"); err != nil {
		logger.Fatal(err)
	}
	defer func() { _ = logger.Close() }()

	// Обработка паники в main
	defer func() {
		if rec := recover(); rec != nil {
			logger.Error(fmt.Sprintf("panic in main: %v", rec))
			logger.Error(fmt.Sprintf("stacktrace:\n%s", debug.Stack()))
			os.Exit(1)
		}
	}()

	logger.Info("app starting")

	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	// Загрузка конфигурации
	cfg, err := env.LoadFromFile("config.env")
	if err != nil {
		logger.Error(fmt.Sprintf("config: %v", err))
		logger.Fatal(err)
	}

	// Подключение к базе данных
	db, err := database.New(cfg, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("database: %v", err))
		logger.Fatal(err)
	}

	// Инициализация зависимостей (repo -> service -> handler)
	walletRepo := repo.NewWalletRepo(db)
	q := queue.NewQueue(walletRepo, cfg.QueueBuffSize, cfg.QueueFlushPeriod)
	go q.ProcessQueue(appCtx)

	walletSrv := service.NewWalletService(q, walletRepo)
	walletHandler := handlers.NewWalletHandler(walletSrv, cfg.RequestTimeout)

	// Rate limiting middleware
	limiter := middleware.NewLimiter(cfg.RateLimit, cfg.RateLimitPeriod)
	server := app.NewServer(walletHandler, limiter)

	// Настройка HTTP сервера
	httpServer := http.Server{
		Addr:         ":8080",
		Handler:      server.Router(),
		ReadTimeout:  time.Minute * 1,
		WriteTimeout: time.Minute * 1,
	}

	logger.Info("server starting on :8080")
	fmt.Println("server listening on :8080")

	// Graceful shutdown: ожидание сигнала завершения
	stopC := make(chan os.Signal, 1)
	signal.Notify(stopC, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(fmt.Sprintf("server error: %v", err))
			logger.Fatal(err)
		}
	}()

	sig := <-stopC
	logger.Info(fmt.Sprintf("shutdown signal received: %v", sig))
	fmt.Printf("shutdown signal received: %v\n", sig)

	appCancel()

	// Завершение работы: закрытие сервера и БД
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error(fmt.Sprintf("shutdown error: %v", err))
	}

	sqlDB, err := db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			logger.Error(fmt.Sprintf("database close error: %v", err))
		}
	}
	logger.Info("graceful shutdown completed")
}
