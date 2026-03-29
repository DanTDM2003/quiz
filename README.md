# Real-time quiz

Design and implementation of a real-time quiz: users join by quiz ID, submit answers, receive consistent score updates, and see a live leaderboard.

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

After `POST /quizzes/{quizId}/participants`, clients open `wss://api.example.com/v1/quizzes/{quizId}/stream` and send a `join` message with the issued `participantId`. The server responds with `welcome` or `error`, then pushes `participant_joined`, `score_updated`, and `leaderboard_updated` for that quiz.

## How to run

Local run commands (Docker, Makefile, service startup) will be added in a later phase. Until then, use this repository as the contract reference for implementation.
