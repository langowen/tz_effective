FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g internal/ports/http/public/docs.go -o internal/ports/http/public/docs

RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/app

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/app .
COPY --from=builder /app/deploy/migrations ./deploy/migrations
COPY --from=builder /app/deploy/config/.env ./.env

EXPOSE 8082

CMD ["./app"]
