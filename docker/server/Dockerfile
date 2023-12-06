FROM golang:1.21 as builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine

# Копируйте скомпилированный бинарный файл из предыдущего шага
COPY --from=builder /app/server /server

EXPOSE 40000

# Запустите ваше приложение
CMD ["/server"]