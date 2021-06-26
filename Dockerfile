FROM alpine as dist
ENV PRIVATE_KEY_PATH=./key/golauth.rsa \
    PUBLIC_KEY_PATH=./key/golauth.rsa.pub

COPY golauth /app/
COPY ./migrations /app/migrations
COPY ./key /app/key
RUN addgroup -S golauth && adduser -S golauth -G golauth \
    && chown -R golauth:golauth  /app
USER golauth
WORKDIR /app
VOLUME /app/key
EXPOSE 8080
CMD ["./golauth"]
