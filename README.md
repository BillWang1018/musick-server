# Musick Server

A TCP-based chat server built with [easytcp](https://github.com/DarthPestilane/easytcp) for real-time messaging with Flutter clients.

## Project Structure

```
musick-server/
├── main.go                     # Entry point - starts the server
├── go.mod                      # Go module dependencies
├── go.sum                      # Dependency checksums
├── client/                     # Go client example for testing
│   └── main.go
└── internal/
    └── app/
        ├── server.go           # Server initialization & route registration
        ├── routes/             # Message route handlers
        │   ├── auth.go         # Authentication routes (login, etc.)
        │   └── echo.go         # Echo test route
        └── services/           # Business logic & external integrations
            ├── session.go      # Session management (user state)
            └── tokenauth.go    # Supabase token verification
```

## Entry Point

**`main.go`** is the application entry point:

1. **Imports** `internal/app` package
2. **Defines** the listen address (`0.0.0.0:5896`)
3. **Creates** a new server instance via `app.New()`
4. **Starts** the server with `server.Run(listenAddr)`

```go
package main

import (
    "log"
    "musick-server/internal/app"
)

const listenAddr = "0.0.0.0:5896"

func main() {
    server := app.New()
    if err := server.Run(listenAddr); err != nil {
        log.Fatalf("server stopped: %v", err)
    }
}
```

## How It Works

### 1. Server Initialization (`internal/app/server.go`)

- **Creates** easytcp server with `DefaultPacker` (length-prefixed framing)
- **Registers hooks**:
  - `OnSessionCreate`: logs client connections
  - `OnSessionClose`: logs disconnections & cleans up session data
- **Calls** `registerRoutes()` to wire message handlers
- **Returns** wrapped server instance

### 2. Route Registration (`internal/app/server.go`)

The `registerRoutes()` function imports and calls registrars from the `routes/` package:

```go
func registerRoutes(s *easytcp.Server) {
    routes.RegisterEchoRoutes(s)   // Test echo handler
    routes.RegisterAuthRoutes(s)   // Login/auth handlers
    // Add more route groups here...
}
```

Each route file exports a `Register*Routes()` function that maps message IDs to handlers.

### 3. Route Handlers (`internal/app/routes/`)

Each handler:
- **Parses** incoming message data (JSON, raw bytes, etc.)
- **Validates** request format
- **Calls** services for business logic
- **Builds** response message
- **Sends** via `ctx.SetResponseMessage()`

Example:
```go
// routes/auth.go
func RegisterAuthRoutes(s *easytcp.Server) {
    s.AddRoute(10, handleLogin)  // Route ID 10 = login
}

func handleLogin(ctx easytcp.Context) {
    // Parse request
    // Verify with Supabase
    // Store session
    // Send response
}
```

### 4. Services (`internal/app/services/`)

Handles non-networking concerns:

- **`session.go`**: Thread-safe user session storage (persists across requests)
- **`tokenauth.go`**: Supabase JWT verification via REST API

Services are called by route handlers to keep them clean and testable.

### 5. Message Flow

```
Flutter Client                   Server
     │                              │
     ├──[route 10, JSON token]─────>│
     │                              ├── routes/auth.go: handleLogin()
     │                              ├── services/tokenauth.go: VerifyToken()
     │                              ├── services/session.go: StoreSession()
     │                              │
     │<────[route 10, response]─────┤
     │                              │
     ├──[route 20, chat msg]───────>│
     │                              ├── services/session.go: IsAuthenticated()
     │                              ├── routes/chat.go: handleSendMessage()
     │                              │
     │<────[route 20, ack]──────────┤
```

## Adding New Features

### Add a new route:

1. **Create** `internal/app/routes/feature.go`
2. **Define** request/response structs
3. **Export** `RegisterFeatureRoutes(s *easytcp.Server)`
4. **Implement** handler functions
5. **Call** from `registerRoutes()` in `server.go`

### Add a new service:

1. **Create** `internal/app/services/feature.go`
2. **Export** public functions for use by routes
3. **Import** in route handlers as needed

## Running

```bash
# Start server
go run main.go

# Test with Go client
go run ./client/main.go

# Test with Flutter client
# (Connect to 0.0.0.0:5896 using Socket.connect)
```

## Message Format

Uses easytcp `DefaultPacker`:

```
┌─────────────┬─────────────┬──────────────────┐
│  dataSize   │     id      │       data       │
│  (4 bytes)  │  (4 bytes)  │   (n bytes)      │
│   uint32    │   uint32    │     []byte       │
└─────────────┴─────────────┴──────────────────┘
```

- **dataSize**: length of `data` field (big-endian)
- **id**: route/message type (big-endian)
- **data**: payload (JSON, raw bytes, etc.)

## Route IDs

Current routes:
- `1`: Echo (test)
- `10`: Login (authentication)

Plan your ID scheme (e.g., 1xxx = auth, 2xxx = chat, 3xxx = presence).
