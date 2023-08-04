STACK_NAME=golauth

prepare:
	cp .env.example .env
	go install github.com/ory/go-acc@latest
	go install github.com/golang/mock/mockgen@v1.6.0
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
	go run main.go

fmt:
	go fmt ./...

test: mock
	go-acc ./...

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o golauth

cover:
	go tool cover -html coverage.txt

mock:
	go generate -v ./...
