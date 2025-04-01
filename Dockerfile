# Этап сборки приложения
FROM golang:1.23.4 AS builder

WORKDIR /app
COPY . .

# Компилируем приложение для Linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o scheduler ./..

# Финальный образ
FROM ubuntu:latest

# Устанавливаем зависимости
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Копируем собранный бинарник и статические файлы
COPY --from=builder /app/scheduler .
COPY web ./web/

# Настройки среды по умолчанию
ENV TODO_PORT=7540
ENV TODO_DBFILE=/data/scheduler.db
ENV TODO_PASSWORD=""

# Открываем порт и указываем точку монтирования для БД
EXPOSE ${TODO_PORT}
VOLUME /data

# Запуск приложения
CMD ["./scheduler"]