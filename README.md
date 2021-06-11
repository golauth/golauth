# golauth

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=golauth_golauth&metric=alert_status)](https://sonarcloud.io/dashboard?id=golauth_golauth)

Authentication server with Golang.

## Usage

Run command with a pre-existing `Postgres` database server:
```
docker run -p 8080:8080 \
    -e DB_HOST=<database_host> \
    -e DB_PORT=<database_port> \
    -e DB_NAME=<database_name> \
    -e DB_USERNAME=<database_username> \
    -e DB_PASSWORD=<database_password>
    golauth/golauth
```

Docker compose example with database creation:

```yaml
version: '3.7'

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

| Env Variable | Description              | Default |
|--------------|--------------------------|---------|
| DB_HOST      | Dabatase hostname        |db|
| DB_PORT      | Database port            |5432|
| DB_NAME      | Database name            |golauth|
| DB_USERNAME  | Database username        |golauth|
| DB_PASSWORD  | Database password        |golauth|
| PORT         | Default application port |8080|

### Accessing

Default user is `admin` and password `admin123`.


```bash
$ curl --request POST \
    --url http://localhost:8080/auth/token \
    --header 'content-type: application/json' \
    --data '{"username": "admin","password": "admin123"}'
```

or 

```bash
$ curl --request POST \
    --url http://localhost:8080/auth/token \
    --header 'content-type: application/x-www-form-urlencoded' \
    --data username=admin \
    --data password=admin123
```