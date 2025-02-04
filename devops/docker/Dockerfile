# 🏗️ Build stage
FROM golang:1.22-alpine AS builder

# 📦 Install build dependencies
RUN apk add --no-cache git make build-base

# 🔧 Set working directory
WORKDIR /app

# 📥 Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# 📂 Copy source code
COPY . .

# 🔨 Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/xyphos ./cmd/main.go

# 🚀 Final stage
FROM gcr.io/distroless/static-debian12:nonroot

# 📝 Labels
LABEL org.opencontainers.image.title="Xyphos"
LABEL org.opencontainers.image.description="Secure multi-tenant key management system"
LABEL org.opencontainers.image.source="https://github.com/theboringhumane/xyphos"
LABEL org.opencontainers.image.vendor="Harsh VARDHAN GOSWAMI"
LABEL org.opencontainers.image.licenses="MIT"

# 📂 Create necessary directories
WORKDIR /app
COPY --from=builder /app/bin/xyphos /app/xyphos

# 💾 Create data directory for BadgerDB
VOLUME ["/app/data"]

# 🔓 Use non-root user
USER nonroot:nonroot

# ⚡ Set entrypoint
ENTRYPOINT ["/app/xyphos"] 