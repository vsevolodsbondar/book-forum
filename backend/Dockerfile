FROM golang:1.27 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /tmp/forum-backend .

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /tmp/forum-backend /app/forum-backend

EXPOSE 8080

CMD ["/app/forum-backend"]
