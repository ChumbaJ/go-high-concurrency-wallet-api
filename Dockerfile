FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o migrate ./cmd/migrate/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/app .
COPY --from=builder /app/migrate .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./app"]
