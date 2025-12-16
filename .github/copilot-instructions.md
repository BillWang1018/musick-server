# Copilot Instructions for Musick Server

## Overview
- TCP chat server using `github.com/DarthPestilane/easytcp` with length-prefixed framing (DefaultPacker: little-endian `dataSize|id|data`; payload is raw bytes/JSON per handler).
- Entry: `main.go` loads `.env` (godotenv), builds server via `internal/app/server.go`, listens on `0.0.0.0:5896`.
- Routes (message IDs):
  - `1` Echo (rejects unauthenticated sessions).
  - `10` Auth: accepts Supabase JWT, verifies via REST, stores session (user ID/email) in-memory.
  - `201` Create room: calls Supabase RPC `create_room_with_owner` with `_owner_id`, `_title`, `_is_private`.
- Session store: `services/session.go` keeps per-connection user info; cleaned on `OnSessionClose`.
- Supabase integration: `SUPABASE_URL`, `SUPABASE_ANON_KEY` from env; JWT passed by client to route 10; room creation currently uses the service role/anon key plus owner_id parameter.

## Key Files
- `main.go`: env load, start server.
- `internal/app/server.go`: easytcp server setup, hooks, route registration.
- `internal/app/routes/`: `auth.go`, `room.go`, `echo.go` define handlers and register IDs.
- `internal/app/services/`: `tokenauth.go` (verify JWT), `session.go` (in-memory sessions), `room.go` (Supabase RPC call).
- `.env` / `.env.example`: required Supabase URL/key.

## Patterns & Conventions
- Routing: register via `s.AddRoute(id, handler)` inside `registerRoutes`; keep handler files under `internal/app/routes` and export `Register*Routes`.
- Handler flow: parse JSON -> validate -> auth check via `services.IsAuthenticated`/`GetSession` -> call service -> respond with `ctx.SetResponseMessage(easytcp.NewMessage(req.ID(), data))`.
- Logging: use `log.Printf` in handlers; connection lifecycle logged in `OnSessionCreate/OnSessionClose`.
- Session data is not persisted across connections; it’s only in-memory per TCP session.
- Supabase HTTP calls set `Authorization: Bearer <token>` (JWT or anon key) and `apikey` header.

## Build & Run
- Prereqs: Go modules, `.env` with `SUPABASE_URL`, `SUPABASE_ANON_KEY`.
- Run server: `go run main.go` (from repo root).
- Sample Go client: `go run ./client/main.go`.

## Supabase Notes
- `tokenauth.VerifyToken` requires a valid JWT from Supabase Auth.
- `CreateRoom` expects the SQL function `create_room_with_owner` to exist (see SQL in README). It passes `_owner_id` explicitly because JWT is not stored in session.
- Ensure RLS/policies allow the service role to call `create_room_with_owner` and that the function inserts the owner into `room_members`.

## When adding features
- Add new route files under `internal/app/routes/*`, export `RegisterXRoutes`, wire in `registerRoutes`.
- Keep request/response structs near handlers; use JSON.
- If a handler needs auth, call `services.IsAuthenticated` and compare against `GetSession` data.
- For Supabase operations, centralize HTTP calls in `internal/app/services/*` and load env via `loadEnv()`.

## Gotchas
- DefaultPacker uses little-endian for `dataSize` and `id`; keep client framing consistent (`dataSize|id|data`).
- Echo route requires authentication (not a public ping).
- Session store currently lacks JWT; if you need user-scoped Supabase calls, pass the token through the auth route and preserve it in session.
