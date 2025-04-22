# ─── Builder Stage ─────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

# Filecoin network. Valid values: mainnet, calibnet
ARG NETWORK="calibnet"

# ensure necessary system dependencies
RUN apk add --no-cache git make

WORKDIR /app

# grab Go modules first to leverage caching
COPY go.mod go.sum ./
RUN go mod download

# copy and compile
COPY cmd/ ./cmd
COPY internal/ ./internal
COPY pkg/ ./pkg
COPY Makefile ./Makefile
COPY version.json ./version.json

RUN make $NETWORK

# ─── Final Stage ──────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache bash

COPY --from=builder /app/storage /usr/local/bin/storage
COPY docker/aggregator-entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
