FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main .

FROM alpine:latest

WORKDIR /AtomicCD

COPY --from=builder /app/main .

EXPOSE 5001

CMD ["./main"]