# ============================================================
# Stage 1: Build the v1claw binary
# ============================================================
FROM golang:1.26.0-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /src

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN make build

# ============================================================
# Stage 2: Minimal runtime image
# ============================================================
FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata curl

RUN addgroup -S v1claw && adduser -S -G v1claw v1claw

ENV V1CLAW_HOME=/home/v1claw/.v1claw

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -fsS http://localhost:18790/health >/dev/null || exit 1

# Copy binary
COPY --from=builder /src/build/v1claw /usr/local/bin/v1claw

RUN mkdir -p /home/v1claw && chown -R v1claw:v1claw /home/v1claw

USER v1claw

# Create v1claw home directory
RUN /usr/local/bin/v1claw onboard

ENTRYPOINT ["v1claw"]
CMD ["gateway"]
