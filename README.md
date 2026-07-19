# microblog

Microblog split into two Go services:

- `api`: REST API, users/posts/likes storage, Kafka event producer.
- `engagement`: Kafka consumer, like counters aggregate, stats REST API.

## Local Infrastructure

Create local env from the example:

```powershell
Copy-Item .env.example .env
```

Run the stack:

```powershell
docker compose -f deployments/docker-compose.yml --env-file .env up --build
```

The compose stack starts:

- `api` on `localhost:8080`
- `engagement` on `localhost:8081`
- `db` Postgres on `localhost:5432`
- `kafka` broker on `localhost:19092`
- `kafka-ui` on `localhost:8082`

Run Goose migrations before using the services:

```powershell
goose -dir db/migrations postgres "$env:DATABASE_URL" up
```

## Events

Events are JSON messages:

```json
{
  "event_id": "uuid",
  "event_type": "PostLiked",
  "occurred_at": "2025-08-12T10:15:30Z",
  "producer": "api",
  "payload": {},
  "trace_id": "optional"
}
```

Topics are configured by env:

- `KAFKA_TOPIC_USER_REGISTERED`
- `KAFKA_TOPIC_POST_CREATED`
- `KAFKA_TOPIC_POST_LIKED`

`engagement` processes `PostLiked` idempotently with `processed_events` and updates `post_like_counters`.

Kafka UI is available at `http://localhost:8082`. Open the `local` cluster, then the Topics page to inspect `user_registered`, `post_created`, and `post_liked` messages.

For local Go runs that use the server infrastructure:

```powershell
$env:DATABASE_URL="postgres://postgres:secret321@188.120.229.55:5432/microblog_db?sslmode=disable"
$env:KAFKA_BROKERS="188.120.229.55:19092"
```

Server Kafka external listener is advertised as `188.120.229.55:19092`.

## Endpoints

API keeps the existing sprint endpoints and adds aliases:

- `POST /register`
- `POST /posts/create`
- `POST /posts`
- `POST /like/{id}?user_id={user_id}`
- `POST /posts/{id}/like?user_id={user_id}`
- `GET /posts/get?id={id}`

Engagement:

- `GET /stats/posts/{id}`
- `GET /healthz`

## Checks

Use a writable Go cache on Windows if the default cache is locked:

```powershell
$env:GOCACHE='C:\projects\microblog\.gocache'
go test ./...
```

Existing testcontainers tests require a supported Docker setup.
