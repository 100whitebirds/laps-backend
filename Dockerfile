FROM golang:1.23-alpine AS builder

RUN apk update && \
    apk add --no-cache gcc musl-dev git

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies (cached if go.mod/go.sum unchanged)
RUN go mod download

# Copy source code
COPY . .

# Build binary with size optimization
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o laps .

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

WORKDIR /app

COPY --from=builder /app/laps .

COPY --from=builder /app/migrations ./migrations

COPY --from=builder /app/docs ./docs

RUN adduser -D -u 1000 appuser && \
    chown -R appuser:appuser /app

USER appuser

ENV GIN_MODE=release

EXPOSE 8080

CMD ["./laps"] 