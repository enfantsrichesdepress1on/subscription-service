FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/subscription-service ./cmd/api

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /bin/subscription-service /app/subscription-service
COPY docs /app/docs
EXPOSE 8080
ENTRYPOINT ["/app/subscription-service"]
