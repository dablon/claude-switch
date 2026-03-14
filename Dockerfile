FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /claude-switch ./cmd/claude-switch

# Final stage
FROM scratch

COPY --from=builder /claude-switch /claude-switch

ENTRYPOINT ["/claude-switch"]
