# Базовый образ для сборки приложения на Go
FROM golang:1.22 AS builder
WORKDIR /app

# Копируем файлы проекта и загружаем зависимости
COPY . .
RUN go mod download

# Компилируем приложение
RUN go build -o main .

# Минимальный образ для запуска
FROM debian:latest
RUN apt-get update && apt-get install -y netcat-openbsd curl
WORKDIR /app
COPY --from=builder /app/main .
COPY wait-for-it.sh /app/wait-for-it.sh

# Делаем wait-for-it.sh исполняемым
RUN chmod +x /app/wait-for-it.sh

# Открываем порт
EXPOSE 8080

# Команда для запуска приложения
CMD ["./wait-for-it.sh", "postgres-db", "5432", "--", "./main"]
