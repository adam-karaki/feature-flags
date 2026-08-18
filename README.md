# Distributed Feature Flag Service

A small Go feature-flag service demonstrating persistent configuration, optimistic concurrency, caching, deterministic percentage evaluation, and event publication.

## Architecture

```text
Client
  |
  v
HTTP API
  |
  +---- PostgreSQL (source of truth)
  |
  +---- Redis (read cache)
  |
  +---- Kafka (change events)
```

## Current implementation

- Create/get/update/delete flags
- Optimistic concurrency using `If-Match` and integer versions
- PostgreSQL persistence
- Redis cache-aside reads
- Kafka change events
- Deterministic percentage rollout based on SHA-256
- Rule operators: `equals`, `not_equals`, `contains`
- JSON API
- Structured request logging
- Graceful HTTP shutdown

## Run locally

Requirements: Go 1.23+, Docker.

```bash
make infra-up
make test
make run
```

The API listens on `:8080` by default.

## API examples

Create:

```bash
curl -X POST localhost:8080/flags \
  -H 'Content-Type: application/json' \
  -d '{"name":"new-checkout","enabled":true,"rules":[{"attribute":"country","operator":"equals","value":"US","percentage":50}]}'
```

Read:

```bash
curl localhost:8080/flags/new-checkout
```

Evaluate:

```bash
curl -X POST localhost:8080/flags/new-checkout/evaluate \
  -H 'Content-Type: application/json' \
  -d '{"subject_id":"user-123","attributes":{"country":"US"}}'
```

Update with optimistic concurrency:

```bash
curl -X PUT localhost:8080/flags/new-checkout \
  -H 'Content-Type: application/json' \
  -H 'If-Match: 1' \
  -d '{"enabled":false,"rules":[]}'
```

## Design notes

PostgreSQL is the source of truth. Redis is a performance optimization; reads fall back to PostgreSQL when the cache misses. Writes use a version predicate so two writers cannot both successfully update the same version.

Kafka events are published after the database/cache mutation. This intentionally keeps the first implementation straightforward; a production implementation should use an outbox pattern to make database state and event publication atomic.

The service is intentionally not a production-ready feature-management platform. Authentication, authorization, multi-tenancy, rate limiting, schema migration tooling, a durable outbox, consumer-side cache propagation, and advanced targeting are outside the initial scope.
