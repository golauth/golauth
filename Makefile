STACK_NAME=golauth
GO111MODULE=on

prepare:
	cp .env.example .env

start-db:
	docker-compose -p ${STACK_NAME} up -d

stop-db:
	docker-compose -p ${STACK_NAME} stop

build-image:
	docker build -t golauth/golauth -f Dockerfile .

run:
	go run main.go

fmt:
	go fmt ./...