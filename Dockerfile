FROM golang:alpine as builder
ENV GO111MODULE=on
WORKDIR /build
COPY . .
RUN apk add --no-cache git \
    && go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o golauth

###
FROM alpine as dist
ENV MIGRATION_SOURCE_URL=./migrations \
    PORT=8080 \
    PRIVATE_KEY_PATH=./key/golauth.rsa \
    PUBLIC_KEY_PATH=./key/golauth.rsa.pub \
    APP_ID=golauth \
    APP_SECRET=15abdaec-735f-423e-96e4-f8f93855cee9 \
    DB_HOST=db \
    DB_PORT=5432 \
    DB_NAME=golauth \
    DB_USERNAME=golauth \
    DB_PASSWORD=golauth

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
