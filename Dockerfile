# Go + lightweight Linux
FROM golang:1.24-alpine

# Working directory inside container
WORKDIR /app

# Required to build with SQLite (CGO)
RUN apk add --no-cache build-base sqlite-dev

# Go dependency files
COPY go.mod go.sum ./

# Download the copied dependencies
RUN go mod download

# Application source code
COPY . .

# Build the Go binary
RUN CGO_ENABLED=1 go build -o myserver .

# Application listens on port 5000
EXPOSE 5000

# Start the server
CMD ["./myserver"]
