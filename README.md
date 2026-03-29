# Real-time quiz

Design and implementation of a real-time quiz: users join by quiz ID, submit answers, receive consistent score updates, and see a live leaderboard.

The service is implemented in **Go** (`go 1.22+`). OpenAPI and JSON Schema contracts in `contracts/` stay language-neutral.

## Problem

- Users join a quiz session using a unique quiz ID; multiple users may join the same session.
- Submitted answers update scores in real time; scoring must stay accurate and consistent across clients.
- A leaderboard shows current standings for the session and updates promptly when scores change.

## Assumptions

- Each quiz has a fixed set of questions; each question has mutually exclusive options and one correct option unless the product rules state otherwise.
- A participant is identified by a server-issued ID after joining; the same user reopening the app may be modeled as a new participant unless session recovery is added later.
- The server is the source of truth for scores and leaderboard order; clients only render server-provided state and events.
- Real-time delivery uses a persistent connection (for example WebSocket) scoped to the quiz session; REST is used for join, submit, and optional snapshot reads.
- Network retries may duplicate submit requests; accept endpoints should support idempotency via a client-supplied idempotency key per answer attempt.

## API and events

- HTTP contract: [contracts/openapi.yaml](contracts/openapi.yaml)
- WebSocket message shapes: [contracts/ws-events.schema.json](contracts/ws-events.schema.json)

After `POST /v1/quizzes/{quizId}/participants`, clients open `wss://api.example.com/v1/quizzes/{quizId}/stream` and send a `join` message with the issued `participantId`. The server responds with `welcome` or `error`, then pushes `participant_joined`, `score_updated`, and `leaderboard_updated` for that quiz.

## Project layout (Go)

| Path | Role |
|------|------|
| `cmd/server` | HTTP entrypoint |
| `internal/domain` | Pure scoring and leaderboard ordering |
| `internal/quiz` | In-memory quiz registry (to be backed by storage later if needed) |
| `internal/api` | HTTP handlers |
| `internal/id` | Small helpers (UUID generation) |
| `contracts/` | OpenAPI + JSON Schema (unchanged by implementation language) |

## Development phases (commit-friendly)

1. **Contracts** — README assumptions + `contracts/` (done).
2. **Domain** — `internal/domain` table-driven tests; no I/O (done).
3. **Join session** — `POST /v1/quizzes/{quizId}/participants` + registry (done).
4. **Submit answers** — `POST /v1/quizzes/{quizId}/answers`, idempotency, score state per participant (done).
5. **Realtime** — WebSocket hub per `quizId`, broadcast `score_updated` / `leaderboard_updated`.
6. **Leaderboard HTTP** — `GET /v1/quizzes/{quizId}/leaderboard` aligned with domain ordering.
7. **Hardening** — Integration tests with multiple goroutines/clients; optional rate limits.
8. **Docker + Makefile** — reproducible local runs on macOS, Linux, and Windows with Docker Desktop.

Each phase should be one focused commit (or `feat` + `test` if you split implementation and tests).

## How to run

Requires Go 1.22 or newer.

```bash
go test ./...
go run ./cmd/server
```

The server listens on port 3000 by default (`PORT` overrides). A seeded quiz id `sample-quiz` exists.

```bash
curl -s -X POST http://localhost:3000/v1/quizzes/sample-quiz/participants \
  -H 'content-type: application/json' \
  -d '{"displayName":"Ada"}'
```

Use the returned `participantId` as `X-Participant-Id` when submitting. Question `q1` accepts options `a`, `b`, or `c` (correct is `a`, 10 points).

```bash
curl -s -X POST http://localhost:3000/v1/quizzes/sample-quiz/answers \
  -H 'content-type: application/json' \
  -H 'X-Participant-Id: <participantId>' \
  -H 'Idempotency-Key: client-req-0001' \
  -d '{"questionId":"q1","selectedOptionId":"a"}'
```

Docker and Makefile are planned for phase 8.
