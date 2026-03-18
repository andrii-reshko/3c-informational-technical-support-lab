# Stage 1: Build
FROM golang:1.21-alpine AS builder
WORKDIR /app

# Копіюємо сирцевий код
COPY main.go .

# Ініціалізуємо модуль "на льоту" та збираємо бінарник
# CGO_ENABLED=0 гарантує статичну збірку, яка точно запуститься в alpine
RUN go mod init stress-server && \
    CGO_ENABLED=0 go build -o stress-server main.go

# Stage 2: Final image
FROM alpine:latest
WORKDIR /root/

# Встановлюємо curl для швидкої перевірки працездатності вручну
RUN apk add --no-cache curl

COPY --from=builder /app/stress-server .

EXPOSE 80
CMD ["./stress-server"]