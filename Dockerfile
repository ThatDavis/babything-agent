# Build stage
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o babything-agent .

# Runtime stage
FROM alpine:3.19
RUN apk add --no-cache ffmpeg ca-certificates
WORKDIR /app
COPY --from=builder /build/babything-agent .
ENTRYPOINT ["./babything-agent"]
