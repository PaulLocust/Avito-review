FROM golang:1.25.2-alpine3.21

WORKDIR /app

# Копируем файлы модулей и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем приложение
RUN go build -o avito-review-service ./cmd/Avito-review/main.go

# Открываем порт и запускаем сервис
EXPOSE 8080
CMD ["./avito-review-service"]