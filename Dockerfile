# Используйте готовый образ с установленным Go
FROM golang:1.18-alpine AS builder

# Определите рабочую директорию в контейнере
WORKDIR /app

# Скопируйте go.mod и go.sum для скачивания зависимостей
COPY go.mod go.sum ./

# Загрузите модули, определенные в go.mod
RUN go mod download

# Скопируйте исходный код в контейнер
COPY . .

# Скопируйте файл .env в контейнер
COPY ./cmd/.env ./cmd/

# Соберите Go-приложение
RUN go build -o /bot-app ./cmd/main.go

# Используйте минимальный образ для запуска
FROM alpine:latest

# Создайте директорию для приложения
WORKDIR /root/

# Скопируйте скомпилированное приложение в рабочую директорию
COPY --from=builder /bot-app .

# Скопируйте файл .env в рабочую директорию контейнера
COPY --from=builder /app/cmd/.env ./

# Определите команду для запуска приложения
CMD ["./bot-app"]
