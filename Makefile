STACK_NAME=golauth
GO111MODULE=on

prepare:
	cp .env.example .env

start-db:
	docker-compose -p ${STACK_NAME} up -d

stop-db:
	docker-compose -p ${STACK_NAME} stop

build-image:
	docker build -t golauth/golauth:dev -f Dockerfile .

run:
	go run main.go

fmt:
	go fmt ./...

test:
	ROOT_PATH=${PWD} go test ./... -coverprofile=coverage.out

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o golauth