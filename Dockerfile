FROM golang:1.24-alpine AS builder

# Установим необходимые утилиты
RUN apk add --no-cache git

# Установим рабочую директорию
WORKDIR /app

# Скопируем go.mod и go.sum для загрузки зависимостей
COPY go.mod go.sum ./

# Загрузим зависимости
RUN go mod download

# Скопируем исходный код
COPY . .

# Соберем бинарный файл
RUN CGO_ENABLED=0 GOOS=linux go build -o pr-reviewer-service ./cmd/api/main.go

# Финальный образ
FROM alpine:latest

# Установим рабочую директорию
WORKDIR /root/

# Скопируем бинарный файл из builder-образа
COPY --from=builder /app/pr-reviewer-service .

# Копируем миграции
COPY migrations/ ./migrations/

# Открываем порт 8080
EXPOSE 8080

# Запускаем приложение
CMD ["./pr-reviewer-service"]