FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init --dir . --output docs
RUN CGO_ENABLED=0 GOOS=linux go build -o server .
RUN for dir in cmd/*/; do \
      name=$(basename "$dir"); \
      CGO_ENABLED=0 GOOS=linux go build -o "bin/$name" "./$dir"; \
    done

FROM alpine:3.21

RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/bin ./bin
COPY --from=builder /app/migrations ./migrations

EXPOSE 9090
CMD ["./server"]
