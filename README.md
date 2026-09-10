# golauth

[![Quality](https://github.com/golauth/golauth/actions/workflows/quality.yaml/badge.svg)](https://github.com/golauth/golauth/actions/workflows/quality.yaml)

---

Simple authentication and authorization server with Golang.

## Usage

Run command with a pre-existing `Postgres` database server:
```
docker run -p 8180:8080 \
    -e DB_HOST=<database_host> \
    -e DB_PORT=<database_port> \
    -e DB_NAME=<database_name> \
    -e DB_USERNAME=<database_username> \
    -e DB_PASSWORD=<database_password> \
    golauth/golauth
```

Docker compose example with database creation:

```yaml
services:
  postgres:
    image: postgres:alpine
    environment:
      - POSTGRES_DB=golauth
      - POSTGRES_USER=golauthuser
      - POSTGRES_PASSWORD=C8HSN2mDvq5Q
    volumes:
      - pgdata:/var/lib/postgresql/data
    networks:
      - golauthnet

  golauth:
    image: golauth/golauth
    links:
      - postgres
    ports:
      - '8180:8080'
    environment:
      - PORT=8080
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=golauth
      - DB_USERNAME=golauthuser
      - DB_PASSWORD=C8HSN2mDvq5Q
    networks:
      - golauthnet

volumes:
  pgdata:

networks:
  golauthnet:
```

##### Environment Variables

| Env Variable             | Description                                                                                                                     |
|--------------------------|---------------------------------------------------------------------------------------------------------------------------------|
| DB_HOST                  | Database hostname                                                                                                               |
| DB_PORT                  | Database port                                                                                                                   |
| DB_NAME                  | Database name                                                                                                                   |
| DB_USERNAME              | Database username                                                                                                               |
| DB_PASSWORD              | Database password (spaces, quotes and backslashes are handled)                                                                  |
| DB_SSLMODE               | TLS mode to Postgres: `disable` (current default), `require`, `verify-ca`, `verify-full`. **The default becomes `require` in a following release** — set it explicitly now. |
| DB_SSLROOTCERT           | PEM bundle to verify the server certificate for `verify-ca` / `verify-full`.                                                     |
| DB_MAX_OPEN_CONNS        | Max open connections in the pool (default 25).                                                                                   |
| DB_MAX_IDLE_CONNS        | Max idle connections kept in the pool (default 25).                                                                              |
| DB_CONN_MAX_LIFETIME     | Max lifetime of a connection, a Go duration (default `30m`).                                                                     |
| DB_CONN_MAX_IDLE_TIME    | Max idle time before a connection is closed, a Go duration (default `5m`).                                                       |
| DB_PING_TIMEOUT          | How long the boot waits for the first connection before failing, a Go duration (default `5s`).                                   |
| RUN_MIGRATIONS           | Run schema migrations at boot (default `true`). Set `false` to run them as a separate job.                                       |
| PORT                     | Application port (default 8080)                                                                                                 |
| CORS_ALLOWED_ORIGINS     | Comma separated browser origins allowed to call the API (default `http://localhost:3000`)                                       |
| APP_ENV                  | When `production`, the process refuses to start without a signing key. Unset or `dev` allows an ephemeral key.                  |
| JWT_PRIVATE_KEY          | RSA private key (PKCS#1 or PKCS#8 PEM, min 2048 bits) used to sign tokens. Inline PEM content.                                  |
| JWT_PRIVATE_KEY_FILE     | Path to a mounted PEM file, used when `JWT_PRIVATE_KEY` is unset. Preferred for Kubernetes secrets.                             |
| JWT_PRIVATE_KEY_PREVIOUS | One or more concatenated PEM blocks kept for verification only, so key rotation is a rolling restart rather than a mass logout. |
| TRUSTED_PROXIES          | Comma separated proxy IPs/CIDRs whose `X-Forwarded-For` is trusted for the client address. Empty means the header is ignored.   |
| LOGIN_RATE_LIMIT         | Requests per window allowed on `POST /auth/token` per client IP (default 10).                                                   |
| LOGIN_RATE_WINDOW        | Sliding window for `LOGIN_RATE_LIMIT`, as a Go duration (default `1m`).                                                         |
| LOGIN_LOCKOUT_THRESHOLD  | Consecutive failed logins before an account is locked (default 5).                                                              |
| LOGIN_LOCKOUT_BASE_DELAY | Lock duration at the threshold; doubles per further failure (default `1m`).                                                     |
| LOGIN_LOCKOUT_MAX_DELAY  | Upper bound on the doubling lock duration (default `15m`).                                                                      |
| PASSWORD_DENYLIST        | `off` (default) disables it; `on` enforces a small embedded common-password list; any other value is a path to a newline-delimited file of forbidden passwords. An unreadable path logs a warning and disables the check. |
| ACCESS_TOKEN_TTL         | Access-token lifetime, a Go duration (default `15m`).                                                                            |
| REFRESH_TOKEN_TTL        | Refresh-token lifetime, a Go duration (default `168h`, i.e. 7 days).                                                             |
| REFRESH_TOKEN_CLEANUP_INTERVAL | How often expired refresh-token rows are purged, a Go duration (default `1h`).                                              |

`CORS_ALLOWED_ORIGINS` no longer defaults to `*`. Set it to the origins of your
front-ends; a wildcard combined with the `authorization` header would let any
site drive the API with a token it obtained from a user.

### Database connection

The connection string is assembled from quoted key/value pairs, so a
`DB_PASSWORD` containing a space, a quote or a backslash works.

TLS is controlled by `DB_SSLMODE`, currently defaulting to `disable`.
**A following release flips the default to `require`**; set `DB_SSLMODE`
explicitly now if your database does not accept TLS, or point `DB_SSLROOTCERT`
at a CA bundle for `verify-full`. The runtime image now ships `ca-certificates`,
so a server certificate from a public CA validates without a bundled root.

The pool is bounded (`DB_MAX_OPEN_CONNS`, default 25) rather than unbounded, so a
traffic spike no longer opens connections until Postgres refuses them. Idle is
kept equal to open by default to avoid reconnect churn under steady load.

Boot pings the database under `DB_PING_TIMEOUT` (default `5s`): an unreachable
database fails the start instead of hanging it. Migrations run at boot unless
`RUN_MIGRATIONS=false`, in which case run them as a separate job before rolling
out the new binary.

### Signing keys and JWKS

The JWT signing key is supplied by configuration so that it is identical across
replicas and survives restarts. Resolution order:

1. `JWT_PRIVATE_KEY` — inline PEM content;
2. `JWT_PRIVATE_KEY_FILE` — path to a mounted PEM file;
3. nothing set and `APP_ENV` unset or `dev` — an ephemeral key is generated and
   a warning is logged. Every restart invalidates all outstanding tokens and a
   second replica cannot verify them;
4. nothing set and `APP_ENV=production` — the process exits without starting.

Generate a key with:

```bash
make gen-key > jwt-key.pem
```

Each key carries a `kid` derived deterministically from its public half, and the
public keys are served, unauthenticated, at:

```
GET /auth/.well-known/jwks.json
```

so any service can verify a token offline.

**Rotation** is a rolling restart: put the new key in `JWT_PRIVATE_KEY` (or its
file), move the old one into `JWT_PRIVATE_KEY_PREVIOUS`, and restart. Tokens
signed by the old key keep validating until they expire; new tokens use the new
key; the JWKS lists both. The first rollout that sets `JWT_PRIVATE_KEY` is a
one-time mass logout — the last one.

### Token introspection

**Offline verification through JWKS is the recommended integration.** A consumer
service fetches `/auth/.well-known/jwks.json` once (cache it, keyed by `kid`),
verifies the access-token signature locally, and reads the claims from the JWT.
No per-request call to golauth, no shared secret.

Two convenience endpoints exist for callers that cannot verify locally:

| Endpoint                | Auth                                  | Success | Body                                                                                  |
|-------------------------|---------------------------------------|---------|---------------------------------------------------------------------------------------|
| `GET /auth/check_token` | bearer in `Authorization`             | `200`   | the verified claims: `sub`, `username`, `firstName`, `lastName`, `authorities`, `exp` |
| `GET /auth/me`          | bearer (standard authenticated route) | `200`   | same shape, read from the token the middleware already verified — no database access  |

A missing or non-bearer `Authorization` header is `400`; an invalid or expired
token is `401` with no claims in the body. `check_token` is public and does
public-key crypto per call, so it is rate limited per client IP with the same
`LOGIN_RATE_LIMIT` / `LOGIN_RATE_WINDOW` budget as the token route.

> **Upgrading:** `check_token` used to answer `204 No Content`. It now answers
> `200` with the claims. A client that only checked for `204` must accept `200`.

### Login throttling

`POST /auth/token` is the credential-stuffing surface and is the only rate
limited route, so an authenticated API under load is unaffected.

- **Per client IP**: a sliding window of `LOGIN_RATE_LIMIT` requests per
  `LOGIN_RATE_WINDOW`. Exceeding it returns `429 Too Many Requests` before the
  request reaches the login logic. The client address is the socket peer unless
  that peer is listed in `TRUSTED_PROXIES`, in which case `X-Forwarded-For` is
  honoured; without that list the header is ignored so it cannot be forged to
  get a fresh bucket.
- **Per account**: after `LOGIN_LOCKOUT_THRESHOLD` consecutive failures the
  account is locked for `LOGIN_LOCKOUT_BASE_DELAY`, doubling on each further
  failure up to `LOGIN_LOCKOUT_MAX_DELAY`. A correct password during the lock
  still fails; a successful login once the lock has expired clears the counter.
  Backed by the `golauth_login_attempt` table.

An unknown username and a wrong password are indistinguishable in status code,
body and response time: the unknown-user path runs the same bcrypt comparison
against a fixed dummy hash.

### Registration and input validation

`POST /auth/signup` validates its body before anything is written. A rejected
payload returns `400 Bad Request` with a `fields` array naming each offending
field:

```json
{ "fields": [ { "field": "password", "message": "must be at least 12 characters" } ] }
```

Rules:

| Field       | Rule                                                                                          |
|-------------|----------------------------------------------------------------------------------------------|
| `username`  | required, 3–50 characters, `a–z 0–9 . _ -` only, stored lower case                            |
| `email`     | required, must parse as an e-mail address, stored lower case                                  |
| `firstName` | required, trimmed, max 255 characters                                                        |
| `lastName`  | required, trimmed, max 255 characters                                                        |
| `document`  | required, trimmed, max 100 characters (the column is `NOT NULL`)                             |
| `password`  | required, **minimum 12 characters**, maximum 72 bytes — bcrypt ignores every byte past 72    |

Optionally, `PASSWORD_DENYLIST` rejects the most common passwords.

The request body no longer accepts an `enabled` field; an account is created
disabled-or-enabled purely by server policy (currently enabled), and activation
is an administrative operation. Sending `enabled` is silently ignored.

A `username` or `email` that already exists returns `409 Conflict`. The response
never contains the password or its hash.

### Token lifecycle and revocation

Login (`POST /auth/token`) returns an OAuth-shaped body:

```json
{ "access_token": "<jwt>", "refresh_token": "<opaque>", "token_type": "Bearer", "expires_in": 900 }
```

- The **access token** is a short-lived JWT (`ACCESS_TOKEN_TTL`, default 15 minutes). It is
  self-contained and is *not* checked against a denylist, so it stays valid until it expires.
- The **refresh token** is a 32-byte opaque string. Only its SHA-256 is stored, in
  `golauth_refresh_token`, so a database dump does not hand over live sessions
  (`REFRESH_TOKEN_TTL`, default 7 days).

Endpoints:

| Endpoint | Auth | Effect |
|---|---|---|
| `POST /auth/token/refresh` | public | Exchange a refresh token for a **new** access + refresh pair. The presented refresh token is rotated: revoked and chained to its successor. Body: `{"refresh_token": "..."}`. |
| `POST /auth/logout` | bearer | Revoke the presented refresh token. Body: `{"refresh_token": "..."}`. Idempotent. |
| `POST /auth/logout/all` | bearer | Revoke every refresh token of the token subject. |

**Rotation and reuse detection.** Every refresh rotates the token. If a refresh token that has
already been rotated away is presented again, that is the signature of a stolen token: golauth
revokes **every** refresh token of that user and logs a `refresh_token_reuse` event. The old
token then buys nothing.

**Deactivation now bites.** Refresh re-checks the user: a disabled account cannot obtain a new
access token, so deactivating a user takes effect within one `ACCESS_TOKEN_TTL` rather than
lasting until the JWT would have expired.

**Logout trade-off.** Logout revokes the refresh token but the current access token keeps working
until it expires. Instant access-token revocation would need a denylist checked on every request;
that is deliberately out of scope. Keep `ACCESS_TOKEN_TTL` short.

**Cleanup.** Expired rows are deleted on start-up and every `REFRESH_TOKEN_CLEANUP_INTERVAL`. A
deployment that prefers an external cron can set the interval very long and run
`DELETE FROM golauth_refresh_token WHERE expires_at < now()` itself.

> **Upgrading:** the access token was 60 minutes and is now 15. Clients that assumed an hour must
> adopt the refresh flow. To migrate gradually, set `ACCESS_TOKEN_TTL=60m` in the deployment,
> ship the refresh support in clients, then drop back to the default.

### Authorization

Only `/auth/token`, `/auth/token/refresh`, `/auth/check_token`, `/auth/signup` and
`/auth/.well-known/jwks.json` are public. Every other endpoint requires a
`Bearer` token, and the role-management endpoints
(`/auth/roles*` and `/auth/users/:id/add-role`) additionally require the `ADMIN`
authority. `GET /auth/users/:id` is available to the user itself or to an admin.

### Accessing

Default user is `admin` and password `admin123`. **Change this password before
exposing the service**: the credential is seeded by the bundled migrations and
is therefore public.

```bash
curl --request POST \
    --url http://localhost:8180/auth/token \
    --header 'content-type: application/json' \
    --data '{"username": "admin","password": "admin123"}'
```

or 

```bash
curl --request POST \
    --url http://localhost:8180/auth/token \
    --header 'content-type: application/x-www-form-urlencoded' \
    --data username=admin \
    --data password=admin123
```

---
