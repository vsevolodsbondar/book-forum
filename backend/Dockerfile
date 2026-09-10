FROM golang:1.27.1-bookworm AS builder

# it is the folder inside of the image that will be built
WORKDIR /app

# copying all files from programs directory and downloading all dependencies
COPY go.mod go.sum ./
RUN go mod download

# copying the result of download operation
COPY . .

#building executable
RUN go build -o /tmp/backend .

# Use the minimal Debian Linux image based on Debian 12 ("Bookworm") as the base image for this stage of the Dockerfile.
FROM debian:bookworm-slim

WORKDIR /app

RUN mkdir -p /app/db

# copy built Linux img with OS
COPY --from=builder /tmp/backend /app/backend

EXPOSE 8080

# running executable using linux os
CMD ["/app/backend"]
