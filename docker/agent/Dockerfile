FROM golang:1.21 as builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o agent ./cmd/agent

FROM alpine

COPY --from=builder /app/agent /agent

CMD ["/agent"]