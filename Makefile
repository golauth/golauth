STACK_NAME=golauth

prepare:
	cp .env.example .env
	go install github.com/ory/go-acc@latest
	go install go.uber.org/mock/mockgen@latest
	go mod download
	go mod tidy

start-db:
	docker-compose -p ${STACK_NAME} up -d

stop-db:
	docker-compose -p ${STACK_NAME} stop

down-db:
	docker-compose -p ${STACK_NAME} down -v

build-image:
	docker build -t golauth/golauth:dev -f Dockerfile .

run:
	go run cmd/api/main.go

fmt:
	go fmt ./...

test: mock
	go-acc --covermode=set -o coverage.txt ./...

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o golauth ./cmd/api/main.go

cover:
	go tool cover -html coverage.txt

mock:
	go generate -v ./...
