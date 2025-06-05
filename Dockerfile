# Stage 1: Build Go API
FROM golang:1.24-alpine AS builder-go

RUN apk add --no-cache git build-base

WORKDIR /app/api

COPY api/go.mod api/go.sum ./
RUN go mod download

COPY api/ ./

RUN go build -o server ./main.go

# Stage 2: Build Vite React app
FROM node:slim-alpine AS builder-web

WORKDIR /app/web

COPY web/package*.json ./
RUN npm install

COPY web/ ./

RUN npm run build

# Stage 3: Final minimal image with nmap
FROM alpine:latest

RUN apk add --no-cache nmap tini

WORKDIR /app

COPY --from=builder-go /app/api/server .
COPY --from=builder-web /app/web/dist ./web/dist

EXPOSE 8080

ENTRYPOINT ["/sbin/tini", "--"]

CMD ["./server"]
