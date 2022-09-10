# golauth

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=golauth_golauth&metric=alert_status)](https://sonarcloud.io/dashboard?id=golauth_golauth)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=golauth_golauth&metric=bugs)](https://sonarcloud.io/dashboard?id=golauth_golauth)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=golauth_golauth&metric=code_smells)](https://sonarcloud.io/dashboard?id=golauth_golauth)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=golauth_golauth&metric=coverage)](https://sonarcloud.io/dashboard?id=golauth_golauth)
[![Lines of Code](https://sonarcloud.io/api/project_badges/measure?project=golauth_golauth&metric=ncloc)](https://sonarcloud.io/dashboard?id=golauth_golauth)

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
version: '3'

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
      - 8180:8080
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

| Env Variable | Description                     |
|--------------|---------------------------------|
| DB_HOST      | Dabatase hostname               |
| DB_PORT      | Database port                   |
| DB_NAME      | Database name                   |
| DB_USERNAME  | Database username               |
| DB_PASSWORD  | Database password               |
| PORT         | Application port (default 8080) |

### Accessing

Default user is `admin` and password `admin123`.


```bash
$ curl --request POST \
    --url http://localhost:8180/auth/token \
    --header 'content-type: application/json' \
    --data '{"username": "admin","password": "admin123"}'
```

or 

```bash
$ curl --request POST \
    --url http://localhost:8180/auth/token \
    --header 'content-type: application/x-www-form-urlencoded' \
    --data username=admin \
    --data password=admin123
```
---
