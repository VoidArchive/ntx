FROM golang:1.25-alpine AS builder

WORKDIR /src

# Copy go.mod and go.sum from apps/api/ directory
# This Dockerfile expects the build context to be the project root
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

# Copy the rest of the application source
COPY apps/api/ .

# Build the application
# CGO_ENABLED=0 for static binary (sqlite pure go implementation via modernc.org/sqlite)
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/ntx ./cmd/ntx

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy compiled binary
COPY --from=builder /bin/ntx /bin/ntx

# Ensure data directory exists for SQLite volume
ENV NTX_DB_PATH=/data/market.db
RUN mkdir -p /data

# Expose port (Railway will set PORT env var, we just expose basic default)
EXPOSE 8080

CMD ["/bin/ntx", "serve"]
