STACK_NAME=golauth

prepare:
	cp .env.example .env
	go install go.uber.org/mock/mockgen@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2
	go mod download
	go mod tidy

start-db:
	docker compose -p ${STACK_NAME} up -d

stop-db:
	docker compose -p ${STACK_NAME} stop

down-db:
	docker compose -p ${STACK_NAME} down -v

build-image:
	docker build -t golauth/golauth:dev -f Dockerfile .

run:
	go run cmd/api/main.go

# gen-key prints a fresh 2048-bit RSA private key in PKCS#8 PEM form. Feed it to
# JWT_PRIVATE_KEY (inline) or JWT_PRIVATE_KEY_FILE (a mounted file).
gen-key:
	@openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048

fmt:
	go fmt ./...

lint: mock
	golangci-lint run ./...

test: mock
	go test -covermode=set -coverpkg=./... -coverprofile=coverage.txt ./...

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o golauth ./cmd/api/main.go

cover:
	go tool cover -html coverage.txt

mock:
	go generate -v ./...
