# Многостадийная сборка для уменьшения размера конечного образа
FROM golang:1.24.0-alpine AS builder

# Установка зависимостей для сборки
RUN apk add --no-cache git ca-certificates

# Установка рабочей директории
WORKDIR /app

# Копирование go модулей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server/main.go

# Вторая стадия: создание минимального образа
FROM scratch

# Копирование сертификатов
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копирование скомпилированного бинарника
COPY --from=builder /app/main .

# Копирование конфигурационных файлов
COPY --from=builder /app/configs/ /configs/

# Указание порта
EXPOSE 8080

# Команда запуска
ENTRYPOINT ["/main"]