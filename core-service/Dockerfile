FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/core-service ./cmd/core-service
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/migrate ./cmd/migrate

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /bin/core-service .
COPY --from=builder /bin/migrate .
COPY --from=builder /app/migrations ./migrations

CMD ["./core-service"]
