# syntax=docker/dockerfile:1

# ---------- builder stage ----------
FROM golang:1.25 AS builder
WORKDIR /src

# Enable Go build and module caches for faster incremental builds
# Note: These mounts require BuildKit; otherwise they are ignored gracefully
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

# Build a static binary for Linux (arch is inferred from the build platform)
ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/budget-go ./cmd/api

# ---------- production stage ----------
# Use a minimal, secure runtime image with CA certificates
FROM gcr.io/distroless/base-debian12:nonroot AS prod
WORKDIR /app

COPY --from=builder /out/budget-go /usr/local/bin/budget-go

# Default port; can be overridden by SERVER_PORT env var
EXPOSE 8001
ENV SERVER_PORT=8001

# Run as non-root user provided by the base image
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/budget-go"]
