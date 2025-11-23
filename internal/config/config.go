package config

import (
	"fmt"
	"os"
)

// Load - загружает конфигурацию приложения
func Load() *Config {
	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		StorageType: getEnv("STORAGE_TYPE", "postgres"), // postgres
		RedisURL:    getEnv("REDIS_URL", ""),
	}
}

// Config - структура конфигурации приложения
type Config struct {
	ServerPort  string // Порт сервера
	DatabaseURL string // Строка подключения к базе данных
	StorageType string // Тип хранилища (in_memory или postgres)
	RedisURL    string // URL Redis
}

// getEnv - возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// GetPostgresDSN возвращает DSN для подключения к PostgreSQL
func (c *Config) GetPostgresDSN() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}

	// По умолчанию формируем строку подключения
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "reviewer_db")

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
}
