# Build stage
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Copy go mod and optional sum files
COPY go.mod go.sum* ./
RUN go mod download

# Copy the source code
COPY . .

# Build the static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kubeguard ./cmd/kubeguard

# Runtime stage
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

# Copy binary from builder
COPY --from=builder /app/kubeguard /kubeguard

# Use nonroot user
USER 65532:65532

ENTRYPOINT ["/kubeguard"]
