FROM golang:1.16-alpine as builder
ENV GO111MODULE=on
WORKDIR /build
COPY . .
RUN apk add --no-cache git make \
    && go mod download \
    && make build

###
FROM alpine as dist
ENV PRIVATE_KEY_PATH=./key/golauth.rsa \
    PUBLIC_KEY_PATH=./key/golauth.rsa.pub \
    MIGRATION_SOURCE_URL=./migrations

COPY --from=builder /build/golauth /app/
COPY --from=builder /build/migrations /app/migrations
COPY --from=builder /build/key /app/key
RUN addgroup -S golauth && adduser -S golauth -G golauth \
    && chown -R golauth:golauth  /app
USER golauth
WORKDIR /app
VOLUME /app/key
EXPOSE 8080
CMD ["./golauth"]
