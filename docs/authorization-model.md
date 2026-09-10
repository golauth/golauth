# Authorization model

Every route is protected by default. `SecurityMiddleware` runs before any
handler and rejects an unauthenticated request unless the path is on its public
allowlist. A route added without a thought is therefore closed, not open — this
is the invariant the regression test in `internal/infra/api/routes_test.go`
(`TestEveryNonPublicRouteRequiresAuthentication`).

Three levels:

| Level             | Meaning                                                           |
|-------------------|-------------------------------------------------------------------|
| **Public**        | No token. On the `SecurityMiddleware` allowlist.                  |
| **Authenticated** | Any valid, unexpired access token. No particular authority.       |
| **ADMIN**         | A valid token **and** the `ADMIN` authority (`RequireAuthority`). |

## Public

| Method & path                     | Purpose                                                                                       |
|-----------------------------------|-----------------------------------------------------------------------------------------------|
| `POST /auth/signup`               | Create an account. Rate-limited input validation; POST only.                                  |
| `POST /auth/token`                | Exchange credentials for an access + refresh pair. Per-IP rate limited and lockout-protected. |
| `POST /auth/token/refresh`        | Rotate a refresh token for a fresh pair.                                                      |
| `GET /auth/check_token`           | Verify a token passed in the `Authorization` header; returns the claims. Per-IP rate limited. |
| `GET /auth/.well-known/jwks.json` | Public signing keys, so any service can verify a token offline.                               |
| `GET /health/live`                | Liveness: the process is up. No dependency checks.                                            |
| `GET /health/ready`               | Readiness: database reachable and signing key loaded.                                         |

## Authenticated (any valid token)

| Method & path           | Purpose                                                      |
|-------------------------|--------------------------------------------------------------|
| `POST /auth/logout`     | Revoke the presented refresh token.                          |
| `POST /auth/logout/all` | Revoke every refresh token of the token subject.             |
| `GET /auth/me`          | Return the caller's own verified claims. No database access. |

## Authenticated, authorized per route

| Method & path                         | Rule                                                                                             |
|---------------------------------------|--------------------------------------------------------------------------------------------------|
| `GET /auth/users/:id`                 | The subject in the token **is** `:id`, **or** the caller has `ADMIN` (`RequireSelfOrAuthority`). |
| `POST /auth/users/:id/add-role`       | `ADMIN`                                                                                          |
| `POST /auth/roles`                    | `ADMIN`                                                                                          |
| `GET /auth/roles/:name`               | `ADMIN`                                                                                          |
| `PUT /auth/roles/:id`                 | `ADMIN`                                                                                          |
| `PATCH /auth/roles/:id/change-status` | `ADMIN`                                                                                          |

## Where the `ADMIN` authority comes from

A user has the `ADMIN` authority when it holds a role that is linked to the
`ADMIN` authority and both the membership and the role are enabled. The seeded
`ADMIN` role carries it. There is no shipped administrator account: the first
one is created at start-up from `BOOTSTRAP_ADMIN_USER` /
`BOOTSTRAP_ADMIN_PASSWORD` when the database has no admin, and the service
refuses to boot if it has neither.

## Keeping this current

The list above mirrors `(*router).Config()` in `internal/infra/api/routes.go`. The
live route table is `app.GetRoutes(true)`; the route-driven tests in
`routes_test.go` iterate it, so an unauthenticated route cannot ship silently
even if this document falls behind.
