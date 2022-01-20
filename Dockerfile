FROM golang:1.17-alpine3.15 as builder
ENV GO111MODULE=on
WORKDIR /build
COPY . .
RUN apk add --no-cache git make \
    && go mod download \
    && make build

###
FROM alpine:3.15 as dist
ENV MIGRATION_SOURCE_URL=./migrations

RUN addgroup -S golauth && adduser -S golauth -G golauth \
    && chown -R golauth:golauth  /app

USER golauth
COPY --from=builder --chown=golauth /build/golauth /app/
COPY --from=builder --chown=golauth /build/ops/migrations /app/migrations
WORKDIR /app
EXPOSE 8080
ENTRYPOINT ["./golauth"]
