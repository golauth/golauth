STACK_NAME=golauth

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
	make mock
	go test ./... -coverprofile=coverage.out

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o golauth

cover:
	go tool cover -html coverage.out

mock:
	go generate -v ./...
