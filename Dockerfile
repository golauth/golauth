FROM golang:1.27-alpine3.24 AS builder
ENV GO111MODULE=on
WORKDIR /build
COPY . .
RUN apk add --no-cache git make \
    && go mod download \
    && make build

###
FROM alpine:3.24 AS dist
ENV MIGRATION_SOURCE_URL=./migrations

# ca-certificates so DB_SSLMODE=require / verify-full can validate a server cert
# signed by a public CA (e.g. RDS, Cloud SQL) without a bundled DB_SSLROOTCERT.
RUN apk add --no-cache ca-certificates \
    && mkdir /app && addgroup -S golauth && adduser -S golauth -G golauth \
    && chown -R golauth:golauth  /app

USER golauth
COPY --from=builder --chown=golauth /build/golauth /app/
COPY --from=builder --chown=golauth /build/migrations /app/migrations
WORKDIR /app
EXPOSE 8080

# The runtime image has no curl or wget, so the binary probes itself: `-healthcheck`
# does a GET against the local /health/live and exits non-zero on failure.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["./golauth", "-healthcheck"]

ENTRYPOINT ["./golauth"]
