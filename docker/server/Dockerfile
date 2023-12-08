FROM golang:1.21 as builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine

COPY --from=builder /app/server /server

CMD ["/server"]