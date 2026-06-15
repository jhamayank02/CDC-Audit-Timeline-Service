# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix netinstall -o server .

# Production stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8000

CMD ["./server"]
