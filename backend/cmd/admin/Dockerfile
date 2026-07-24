# ── Admin Service: Multi-stage Docker build ──────────────────────────────
# Stage 1: Compile the Go binary
FROM docker.m.daocloud.io/library/golang:1.25-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct

RUN apk --no-cache add ca-certificates git

WORKDIR /app

# Copy dependency files first for better layer caching.
COPY go.mod go.sum ./
# Strip host-specific replace directives (they reference absolute host paths)
# and lower the go version directive to match the available toolchain.
# The project uses Go 1.22 language features; go 1.26.0 in go.mod is just a minimum.
RUN sed -i 's/^go 1\.26\.0/go 1.25.0/' go.mod && \
    sed -i '/^replace (/,/^)/d' go.mod
RUN go mod download

# Copy the rest of the source.
COPY . .

# Re-strip (COPY overwrote go.mod) and tidy go.sum entries.
# go.sum entries from the host go.mod reference replace-target paths;
# after stripping the replace block we need fresh checksums for the
# real module versions from the registry.
RUN sed -i 's/^go 1\.26\.0/go 1.25.0/' go.mod && \
    sed -i '/^replace (/,/^)/d' go.mod && \
    go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOTOOLCHAIN=local \
    go build -ldflags="-s -w" -o /admin ./backend/cmd/admin

# ── Stage 2: Minimal runtime ─────────────────────────────────────────────
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user for security.
RUN addgroup -S wms && adduser -S wms -G wms

COPY --from=builder /admin /usr/local/bin/admin
COPY migrations/ /migrations/

USER wms

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=5s --retries=3 \
    CMD curl -sf http://localhost:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/admin"]
