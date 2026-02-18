# Этап сборки
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копируем go.mod
COPY go.mod .

# Копируем исходный код
COPY app/ .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o screenguru .

# Финальный образ
FROM alpine:latest

WORKDIR /app

# Создаем директорию для данных
RUN mkdir -p /data

# Копируем бинарник из этапа сборки
COPY --from=builder /app/screenguru .

# Копируем шаблоны
COPY --from=builder /app/templates ./templates

# Копируем статические файлы
COPY --from=builder /app/templates/static ./templates/static

# Копируем ченджлог
COPY changelog.md .

# Открываем порт
EXPOSE 8000

# Запускаем приложение
CMD ["./screenguru"]
