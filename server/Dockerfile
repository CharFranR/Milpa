FROM golang:1.26-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/api ./cmd/api && \
    go build -ldflags="-w -s" -o /app/migrate ./cmd/migrate


FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/api .
COPY --from=builder /app/migrate .
COPY --from=builder /app/infrastructure/adapters/secondary/repository/migrations ./infrastructure/adapters/secondary/repository/migrations

EXPOSE 8080
