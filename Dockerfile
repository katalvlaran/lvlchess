# Dockerfile.go
#
# 1) Stage builder на базе golang:1.19-alpine (можно другую версию, лишь бы поддерживала Go)
# 2) Копируем go.mod, go.sum, качаем зависимости
# 3) Копируем всё остальное, собираем бинарь
# 4) Финальный образ: минимальный alpine, копируем бинарник, выставляем порт

FROM golang:1.23-alpine AS builder

WORKDIR /app

# Скопируем go.mod и go.sum заранее
COPY go.mod go.sum ./
RUN go mod download

# Скопируем весь проект lvlchess
COPY . .

# Сборка
RUN go build -o telega-chess ./cmd/bot.go


# Final stage
FROM alpine:3.17
WORKDIR /app

# Скопировать наш скомпилённый бинарник
COPY --from=builder /app/telega-chess /app/

# порт 8080 (как указано в docker-compose)
EXPOSE 8080

# Запускаем приложение
CMD ["/app/telega-chess"]
