STACK_NAME=golauth

# `?=` so the environment wins: the Dockerfile builder passes buildx's
# TARGETOS/TARGETARCH through to cross-compile, while a bare `make build` still
# produces the linux/amd64 binary it always has.
GOOS ?= linux
GOARCH ?= amd64

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

# No --target: the last stage in the Dockerfile is the default (distroless),
# which is what the pipeline publishes as `latest`. Pass VARIANT to get another:
#   make build-image VARIANT=alpine
VARIANT ?=
build-image:
	docker build $(if $(VARIANT),--target dist-$(VARIANT),) \
		-t golauth/golauth:dev$(if $(VARIANT),-$(VARIANT),) -f Dockerfile .

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
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-w -s" -o golauth ./cmd/api/main.go

cover:
	go tool cover -html coverage.txt

mock:
	go generate -v ./...
