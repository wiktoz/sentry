FROM alpine:latest
RUN apk add --no-cache nmap nmap-scripts tini bash curl
WORKDIR /app
COPY server .
COPY dist ./web/dist
EXPOSE 8080
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["./server"]