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

The integration tests under `pkg/infra/repository/postgres` and `tests/` start
throwaway Postgres containers with testcontainers, so Docker must be running.

CI also runs `govulncheck ./...` and enforces a coverage floor; keep both green.

## House rules

### Layering

The dependency rule is `domain → application → infra`; an inner layer must not
import an outer one. `pkg/application` must not import `pkg/infra` at all — this
is enforced by `TestApplicationDoesNotImportInfra` in
`pkg/application/layering_test.go`. The route-level authorization model is in
[`docs/authorization-model.md`](docs/authorization-model.md).

### Constructors

A `NewX` constructor returns the interface `X`, not the unexported concrete
type. Callers depend on the interface; the struct stays private.

### File names

One file per principal type or use case, named after it. Two conventions are in
use and each package must pick one and stay uniform:

- `pkg/domain/entity` uses **lowerCamelCase** after the type: `refreshToken.go`,
  `loginAttempt.go`, `user.go`.
- Everywhere else uses **PascalCase** matching the type or use case:
  `RefreshAccessToken.go`, `UserRepository.go`, `HealthController.go`.

Do not mix them within a package. When in doubt, match the files already there.

### The domain layer

Entities may carry small, pure behaviours that name an invariant
(`func (u *User) IsActive() bool`). Request/response shapes and transport
concerns belong in `pkg/infra/api/controller/model`, not in `entity`.

## `pkg/` versus `internal/`

**Decision: the tree stays under `pkg/` for now.**

golauth is a deployable service (`cmd/api`), not a library, and nothing outside
this module imports it. Go's `internal/` would compiler-enforce that and make
exposing a public API a deliberate act rather than an accident, which is the
better end state. It is not done yet only because the rename touches every
import in ~80 files and would swamp any change it rides along with; it should
land as its own isolated PR:

```sh
git mv pkg internal
grep -rl 'golauth/golauth/pkg/' --include='*.go' . \
  | xargs sed -i 's#golauth/golauth/pkg/#golauth/golauth/internal/#g'
# update the two prefixes in pkg/application/layering_test.go
make mock && make test
```

Until then, treat everything under `pkg/` as private to this module. There is no
stable public API.

`tests/postgrescontainer.go` is a non-test helper package sitting next to the
root integration tests. When the move above happens it should become
`internal/testsupport` so a consumer of the module cannot import it.
