FROM golang:1.20-alpine3.17 AS builder
ENV GO111MODULE=on
WORKDIR /build
COPY . .
RUN apk add --no-cache git make \
    && go mod download \
    && make build

###
FROM alpine:3.17 AS dist
ENV MIGRATION_SOURCE_URL=./migrations

RUN mkdir /app && addgroup -S golauth && adduser -S golauth -G golauth \
    && chown -R golauth:golauth  /app

USER golauth
COPY --from=builder --chown=golauth /build/golauth /app/
COPY --from=builder --chown=golauth /build/migrations /app/migrations
WORKDIR /app
EXPOSE 8080
ENTRYPOINT ["./golauth"]
