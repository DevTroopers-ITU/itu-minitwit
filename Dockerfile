# Stage 1 - Builder (has build tools, binutils etc. - thrown away after)
FROM golang:1.25-alpine AS builder
WORKDIR /app
# hadolint ignore=DL3018
RUN apk add --no-cache build-base sqlite-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o myserver .

# Stage 2 - Final image (only the compiled binary, no build tools)
FROM alpine:3.23
WORKDIR /app
# hadolint ignore=DL3018
RUN apk upgrade --no-cache
COPY --from=builder /app/myserver .
RUN addgroup -S appgroup && adduser -S appuser -G appgroup \
    && chown -R appuser:appgroup /app
USER appuser
EXPOSE 8080
CMD ["./myserver"]