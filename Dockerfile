FROM golang:1.16-alpine as builder
ENV GO111MODULE=on
WORKDIR /build
COPY . .
RUN apk add --no-cache git make \
    && go mod download \
    && make build

###
FROM alpine as dist
ENV MIGRATION_SOURCE_URL=./migrations

COPY --from=builder /build/golauth /app/
COPY --from=builder /build/ops/migrations /app/migrations
RUN addgroup -S golauth && adduser -S golauth -G golauth \
    && chown -R golauth:golauth  /app
USER golauth
WORKDIR /app
EXPOSE 8080
CMD ["./golauth"]
