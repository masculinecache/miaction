# Stage 1: Build Go backend
FROM golang:1.25-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Stage 2: Final minimal image
FROM alpine:3.19
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=go-builder /app/server ./
COPY templates/ ./templates/
EXPOSE 8080
USER nobody
CMD ["./server"]
