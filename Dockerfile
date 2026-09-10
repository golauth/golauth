# Three runtime variants share one builder. The stage order matters: the last
# stage is what a bare `docker build .` produces, so dist-distroless sits at the
# bottom because it is the default the pipeline publishes as `latest`.
#
#   docker build -t golauth .                          # distroless (default)
#   docker build --target dist-alpine -t golauth .     # alpine
#   docker build --target dist-debian -t golauth .     # debian
#
# The ENV / WORKDIR / EXPOSE / HEALTHCHECK / ENTRYPOINT block is repeated in each
# runtime stage. Docker has no way to share it across different bases; the three
# copies must stay identical, and the boot-and-go-healthy test for every variant
# is what catches it when they do not.

# ---------------------------------------------------------------------------
# builder
# ---------------------------------------------------------------------------
# Pinned to the *build* platform and cross-compiling to the target. CGO_ENABLED=0
# makes a Go cross-build free; letting buildx emulate arm64 under QEMU instead
# would cost three to four times the whole build.
FROM --platform=$BUILDPLATFORM golang:1.27-alpine3.24 AS builder
ENV GO111MODULE=on
WORKDIR /build
COPY . .

# Supplied by buildx from its --platform list. The Makefile defaults them to
# linux/amd64, so a bare `make build` outside Docker behaves as it always has.
ARG TARGETOS
ARG TARGETARCH
RUN apk add --no-cache git make \
    && go mod download \
    && GOOS=${TARGETOS} GOARCH=${TARGETARCH} make build

# ---------------------------------------------------------------------------
# dist-alpine -- published as :X.Y.Z-alpine
# ---------------------------------------------------------------------------
# What `latest` was before the distroless switch. Kept as the migration path for
# anyone who needs a shell in the container or uses this image as a base.
FROM alpine:3.24 AS dist-alpine
ENV MIGRATION_SOURCE_URL=./migrations

# apk upgrade pulls the current Alpine security patches (the `3.24` tag lags the
# repository, so a fixed OpenSSL etc. would otherwise ship stale and fail the
# image scan). ca-certificates lets DB_SSLMODE=require / verify-full validate a
# server cert signed by a public CA (e.g. RDS, Cloud SQL) without a bundled
# DB_SSLROOTCERT.
RUN apk upgrade --no-cache \
    && apk add --no-cache ca-certificates \
    && mkdir /app && addgroup -S golauth && adduser -S golauth -G golauth \
    && chown -R golauth:golauth  /app

USER golauth
COPY --from=builder --chown=golauth /build/golauth /app/
COPY --from=builder --chown=golauth /build/migrations /app/migrations
WORKDIR /app
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["./golauth", "-healthcheck"]
ENTRYPOINT ["./golauth"]

# ---------------------------------------------------------------------------
# dist-debian -- published as :X.Y.Z-debian
# ---------------------------------------------------------------------------
# For estates whose security team only certifies glibc/Debian bases.
FROM debian:13-slim AS dist-debian
ENV MIGRATION_SOURCE_URL=./migrations

# apt-get upgrade is here for the same reason the Alpine stage runs apk upgrade:
# the slim tag lags its security archive, and the Trivy gate fails on a CVE that
# already has a fix. One layer, and the apt lists are dropped so they do not ship.
RUN apt-get update \
    && apt-get upgrade -y \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir /app \
    && groupadd -r golauth && useradd -r -g golauth -d /app golauth \
    && chown -R golauth:golauth /app

USER golauth
COPY --from=builder --chown=golauth:golauth /build/golauth /app/
COPY --from=builder --chown=golauth:golauth /build/migrations /app/migrations
WORKDIR /app
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["./golauth", "-healthcheck"]
ENTRYPOINT ["./golauth"]

# ---------------------------------------------------------------------------
# dist-distroless -- the default, published as :latest and :X.Y.Z
# ---------------------------------------------------------------------------
# No shell, no package manager, no libc surface: the smallest thing that can run
# a static Go binary. This is only possible because the binary probes itself --
# `-healthcheck` does a GET against the local /health/live -- so the runtime
# image needs neither curl nor a shell for the HEALTHCHECK below. Keep that
# CMD in exec form; the shell form would need a /bin/sh that does not exist here.
#
# Pinned by digest, not just by tag, because `:nonroot` carries no version at
# all -- without the digest there is no way to say what was shipped. It also has
# no package manager, so unlike the two stages above it cannot patch itself at
# build time: a new digest is the *only* way this variant gets a security fix.
# The tag is kept alongside so Dependabot's docker ecosystem can track and bump it.
FROM gcr.io/distroless/static-debian13:nonroot@sha256:1c2c046bc09ed40fad370b599a0b1ae7987f55b01e247cf27a7c27cd97e5bbc7 AS dist-distroless
ENV MIGRATION_SOURCE_URL=./migrations

# No RUN block: the base already provides ca-certificates, tzdata, and the
# nonroot user (uid 65532) in /etc/passwd. There is no shell to run one with.
WORKDIR /app
COPY --from=builder --chown=nonroot:nonroot /build/golauth /app/
COPY --from=builder --chown=nonroot:nonroot /build/migrations /app/migrations
USER nonroot:nonroot
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["./golauth", "-healthcheck"]
ENTRYPOINT ["./golauth"]
