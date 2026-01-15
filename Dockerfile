# Builder
FROM golang:1.20-alpine AS builder
WORKDIR /src

# Install git and ca-certificates so `go mod download` works in the builder
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /bin/rate-engine ./cmd/api

# Final runtime image
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /bin/rate-engine /bin/rate-engine
EXPOSE 8080
ENTRYPOINT ["/bin/rate-engine"]
