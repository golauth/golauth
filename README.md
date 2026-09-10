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
| DB_PASSWORD              | Database password                                                                                                               |
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

`CORS_ALLOWED_ORIGINS` no longer defaults to `*`. Set it to the origins of your
front-ends; a wildcard combined with the `authorization` header would let any
site drive the API with a token it obtained from a user.

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

### Authorization

Only `/auth/token`, `/auth/check_token`, `/auth/signup` and
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
