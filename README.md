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
