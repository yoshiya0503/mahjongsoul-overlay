# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Production stage
FROM alpine:latest AS production
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/main .
EXPOSE 8787
CMD ["./main"]
