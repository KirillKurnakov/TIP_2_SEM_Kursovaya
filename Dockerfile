# Этап сборки: используем официальный образ Go
FROM golang:1.24.3-alpine AS builder

# Установим рабочую директорию
WORKDIR /app

# Проверим, что Go установлен (для отладки)
RUN go version

# Скопируем зависимости Go-модуля
COPY go.mod go.sum ./
RUN go mod download

# Скопируем остальные файлы проекта в контейнер
COPY . .

# Скомпилируем приложение (выходной файл main)
RUN go build -o main .

# Этап продакшн: создаём минимальный образ для работы приложения
FROM alpine:latest

# Устанавливаем необходимые зависимости (например, для работы с TLS)
RUN apk --no-cache add ca-certificates

# Копируем скомпилированный бинарник из предыдущего этапа
COPY --from=builder /app/main /app/main

# Устанавливаем рабочую директорию для приложения
WORKDIR /app

# Устанавливаем переменные окружения для базы данных (можно переопределить через docker-compose)
ENV DB_HOST=db
ENV DB_USER=postgres
ENV DB_PASSWORD=postgres
ENV DB_NAME=tipOnlineShop
ENV DB_PORT=5432

# Открываем порт 8080 для приложения
EXPOSE 8080

# Запускаем приложение
CMD ["./main"]