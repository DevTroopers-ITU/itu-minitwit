# Go + lightweight Linux
FROM golang:1.25-alpine

# Working directory inside container
WORKDIR /app

# Required to build with SQLite (CGO)
# hadolint ignore=DL3018
# upgrade all Alpine OS packages to their latest patched versions to fix known CVEs.
RUN apk add --no-cache build-base sqlite-dev && apk upgrade --no-cache

# Go dependency files
COPY go.mod go.sum ./

# Download the copied dependencies
RUN go mod download

# Application source code
COPY . .

# Build the Go binary
RUN CGO_ENABLED=1 go build -o myserver .

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Give the new user ownership of the app directory
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Application listens on port 8080
EXPOSE 8080

# Start the server
CMD ["./myserver"]
