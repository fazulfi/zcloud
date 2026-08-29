# =============================================================================
# Sub2API Multi-Stage Dockerfile
# =============================================================================
# Stage 1: Build frontend (Vue/Vite)
# Stage 2: Build Go backend (with embedded frontend, -tags embed)
# Stage 3: Final minimal image
# =============================================================================

ARG GOLANG_IMAGE=golang:1.27.0-alpine
ARG ALPINE_IMAGE=alpine:3.20
ARG GOPROXY=https://goproxy.cn,direct
ARG GOSUMDB=sum.golang.google.cn

# -----------------------------------------------------------------------------
# Stage 1: Frontend (prebuilt dist). The SPA is built on the host (pnpm build)
# with the target Vite mode (per environment) so VITE_API_BASE_URL matches
# the environment, then embedded into the Go binary via -tags embed.
# -----------------------------------------------------------------------------
FROM ${ALPINE_IMAGE} AS frontend-builder

WORKDIR /app/frontend
COPY frontend/dist ./dist

# -----------------------------------------------------------------------------
# Stage 2: Backend Builder
# -----------------------------------------------------------------------------
FROM ${GOLANG_IMAGE} AS backend-builder

# Build arguments for version info (set by CI)
ARG VERSION=
ARG COMMIT=docker
ARG DATE
ARG GOPROXY
ARG GOSUMDB

ENV GOPROXY=${GOPROXY}
ENV GOSUMDB=${GOSUMDB}

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app/backend

# Copy go mod files first (better caching)
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source first
COPY backend/ ./

# Copy prebuilt frontend for -tags embed (//go:embed all:dist in internal/web)
# Built in Stage 1 (frontend-builder) with FRONTEND_MODE so the SPA's
# VITE_API_BASE_URL matches the target environment (per environment).
COPY --from=frontend-builder /app/frontend/dist ./internal/web/dist

# Build the backend binary (BuildType=release for CI builds).
# The frontend is embedded into the binary (-tags embed) so app/api/root
# all serve from 127.0.0.1:18080 (embedded frontend decision, no Vercel).
# Version precedence: build arg VERSION > exact git tag > cmd/server/VERSION
RUN VERSION_VALUE="${VERSION}" && \
    if [ -z "${VERSION_VALUE}" ]; then VERSION_VALUE="$(./scripts/resolve-version.sh)"; fi && \
    DATE_VALUE="${DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}" && \
     CGO_ENABLED=0 GOOS=linux go build -tags embed \
     -ldflags="-s -w -X main.Version=${VERSION_VALUE} -X main.Commit=${COMMIT} -X main.Date=${DATE_VALUE} -X main.BuildType=release" \
    -o /app/sub2api \
    ./cmd/server

# -----------------------------------------------------------------------------
# Final Runtime Image
# -----------------------------------------------------------------------------
FROM ${ALPINE_IMAGE}

# Labels
LABEL maintainer="Wei-Shaw <github.com/Wei-Shaw>"
LABEL description="Sub2API - AI API Gateway Platform"
LABEL org.opencontainers.image.source="https://github.com/Wei-Shaw/sub2api"

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    su-exec \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1000 sub2api && \
    adduser -u 1000 -G sub2api -s /bin/sh -D sub2api

# Set working directory
WORKDIR /app

# Copy binary and runtime fallback resources from builder
COPY --from=backend-builder /app/sub2api /app/sub2api
COPY --from=backend-builder --chown=sub2api:sub2api /app/backend/resources /app/resources

# Create data directory
RUN mkdir -p /app/data && chown -R sub2api:sub2api /app

# Copy entrypoint script (fixes volume permissions then drops to sub2api)
COPY deploy/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

# Expose port (can be overridden by SERVER_PORT env var)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget -q -T 5 -O /dev/null http://localhost:${SERVER_PORT:-8080}/health || exit 1

# Run the application (entrypoint fixes /app/data ownership then execs as sub2api)
ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/sub2api"]
