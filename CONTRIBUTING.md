# Contributing to golauth

Thanks for helping. This is an authentication server, so correctness and a
reviewable history matter more than speed.

The community guidelines live in [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md).
Security issues go through the process in [`SECURITY.md`](SECURITY.md), never a
public issue or PR.

## Getting set up

```sh
make prepare     # installs mockgen + golangci-lint, downloads modules
make start-db    # local Postgres via docker compose
```

## Before you open a PR

```sh
make mock        # regenerate mocks (they are gitignored, never committed)
make fmt
make lint        # golangci-lint, must be clean
make test        # go test ./... with the coverage profile
```

The integration tests under `internal/infra/repository/postgres` and `tests/`
start throwaway Postgres containers with testcontainers, so Docker must be
running.

CI also runs `govulncheck ./...` and enforces a coverage floor; keep both green.

## House rules

### Layering

The dependency rule is `domain → application → infra`; an inner layer must not
import an outer one. `internal/application` must not import `internal/infra` at
all — this is enforced by `TestApplicationDoesNotImportInfra` in
`internal/application/layering_test.go`. The route-level authorization model is
in [`docs/authorization-model.md`](docs/authorization-model.md).

### Constructors

A `NewX` constructor returns the interface `X`, not the unexported concrete
type. Callers depend on the interface; the struct stays private.

### File names

One file per principal type or use case, named after it. Two conventions are in
use and each package must pick one and stay uniform:

- `internal/domain/entity` uses **lowerCamelCase** after the type:
  `refreshToken.go`, `loginAttempt.go`, `user.go`.
- Everywhere else uses **PascalCase** matching the type or use case:
  `RefreshAccessToken.go`, `UserRepository.go`, `HealthController.go`.

Do not mix them within a package. When in doubt, match the files already there.

### The domain layer

Entities may carry small, pure behaviours that name an invariant
(`func (u *User) IsActive() bool`). Request/response shapes and transport
concerns belong in `internal/infra/api/controller/model`, not in `entity`.

## `internal/`, not `pkg/`

golauth is a deployable service (`cmd/api`), not a library, and nothing outside
this module imports it. The whole tree therefore lives under `internal/`, so the
Go toolchain refuses any import of it from another module: exposing a public API
has to be a deliberate move of code out of `internal/`, never an accident.

There is no stable public API. Do not add a `pkg/` directory back without a
concrete external consumer and a versioning commitment to go with it.

`internal/testsupport` (the testcontainers Postgres helper) is under `internal/`
for the same reason — it is wiring for this module's own tests, not something a
consumer should be able to pull in.
