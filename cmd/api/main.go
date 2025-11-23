package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"pr-reviewer-assignment-service/internal/cache"
	"pr-reviewer-assignment-service/internal/config"
	"pr-reviewer-assignment-service/internal/handlers"
	"pr-reviewer-assignment-service/internal/interfaces"
	"pr-reviewer-assignment-service/internal/metrics"
	"pr-reviewer-assignment-service/internal/postgres"
	"pr-reviewer-assignment-service/internal/services"
)

func main() {
	// Настройка логирования
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	logger.Info("Запуск сервиса назначения ревьюеров для PR")

	// Загрузка конфигурации
	cfg := config.Load()

	var store interfaces.Repository
	var cacheClient *cache.RedisCache
	var closers []func() error

	// Создаем PostgreSQL хранилище
	// Пробуем подключиться к базе данных
	db, err := sql.Open("postgres", cfg.GetPostgresDSN())
	if err != nil {
		logger.Error("Ошибка подключения к PostgreSQL", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Проверим подключение перед выполнением миграций
	if err := db.Ping(); err != nil {
		logger.Error("Ошибка пинга PostgreSQL", slog.String("error", err.Error()))
		_ = db.Close()
		os.Exit(1)
	}

	// Выполняем миграции
	if err := postgres.Migrate(db, logger); err != nil {
		logger.Error("Ошибка выполнения миграций", slog.String("error", err.Error()))
		_ = db.Close()
		os.Exit(1)
	}

	dbStorage, err := postgres.NewPostgresStorage(cfg.GetPostgresDSN(), logger)
	if err != nil {
		logger.Error("Ошибка создания PostgreSQL хранилища", slog.String("error", err.Error()))
		_ = db.Close()
		os.Exit(1)
	}

	store = dbStorage
	closers = append(closers, func() error {
		_ = dbStorage.Close()
		return db.Close()
	})

	// Создание Redis кэша, если URL указан
	if cfg.RedisURL != "" {
		redisCache, err := cache.NewRedisCache(cfg.RedisURL, logger)
		if err != nil {
			logger.Error("Ошибка подключения к Redis", slog.String("error", err.Error()))
		} else {
			cacheClient = redisCache
			closers = append(closers, redisCache.Close)
		}
	}

	// Создание сервиса
	reviewerService := services.NewReviewerService(store, cacheClient, nil, logger)

	// Создание обработчиков
	handler := handlers.NewHandler(reviewerService, logger)

	// Настройка маршрутов
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(metrics.HTTPMetricsMiddleware)

	// Добавляем маршрут для метрик Prometheus
	r.Handle("/metrics", promhttp.Handler())

	// Регистрация маршрутов
	handler.RegisterRoutes(r)

	logger.Info("Сервер запускается на порту", slog.String("port", cfg.ServerPort))

	// Создание HTTP-сервера
	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	// Запуск сервера в отдельной горутине
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Ошибка при запуске сервера", slog.String("error", err.Error()))
		}
	}()

	logger.Info("Сервис успешно запущен")

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Получен сигнал завершения, остановка сервера...")

	// Плавное завершение работы HTTP-сервера
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Ошибка при завершении работы сервера", slog.String("error", err.Error()))
	}

	// Закрываем все соединения
	for _, closer := range closers {
		if err := closer(); err != nil {
			logger.Error("Ошибка при закрытии ресурса", slog.String("error", err.Error()))
		}
	}

	logger.Info("Сервер успешно остановлен")
}
