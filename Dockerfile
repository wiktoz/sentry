# Stage 1: Build Go API (native on Mac ARM64)
FROM golang:1.24-alpine AS builder-go
RUN apk add --no-cache git build-base
WORKDIR /app/api
COPY api/go.mod api/go.sum ./
RUN go mod download
COPY api/ ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./main.go

# Stage 2: Build Vite React app (native node, not arm32v7!)
FROM node:24-alpine AS builder-web
WORKDIR /app/web
COPY web/package*.json ./
RUN npm install
COPY web/ ./
RUN npm run build

# Stage 3: Final lightweight ARMv7 runtime
FROM alpine:latest

# Install tools MikroTik needs
RUN apk add --no-cache nmap nmap-scripts tini bash curl

WORKDIR /app

# Copy in binaries and frontend from native builds
COPY --from=builder-go /app/api/server .
COPY --from=builder-web /app/web/dist ./web/dist

EXPOSE 8080

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["./server"]
