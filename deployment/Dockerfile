# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix netinstall -o api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix netinstall -o consumer ./cmd/consumer

# Production stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/api .
COPY --from=builder /app/consumer .

EXPOSE 8000

CMD ["./api"]
