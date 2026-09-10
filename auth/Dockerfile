FROM golang:1.27.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o auth .

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/auth .

EXPOSE 8081

CMD ["./auth"]
