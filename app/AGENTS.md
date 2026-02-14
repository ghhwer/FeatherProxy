# FeatherProxy App — Structure & Design

This document describes the **overall structure and build pattern** of the `app` package so AI assistants can follow existing conventions.

## Directory layout

```
app/
├── main.go                 # Entry: env, DB, migrate, repo, server, Run(ctx)
├── go.mod / go.sum
├── .env / .env.example
└── internal/
    ├── database/
    │   ├── handler.go      # DB connection (GORM), config from env, AutoMigrate, Close
    │   ├── repository.go   # Repository type alias, ErrProtocolMismatch re-export, NewRepository constructor
    │   ├── repo/           # Repository interface and ErrProtocolMismatch (no impl deps)
    │   ├── impl/           # Concrete repository implementation (GORM, objects, schema conversions)
    │   ├── token/          # EncryptToken / DecryptToken for auth tokens (used by impl)
    │   ├── schema/         # Domain / API types (no ORM; JSON-friendly)
    │   │   └── <ENTITY>.go
    │   └── objects/        # ORM entities (GORM tags, table names); schema ↔ object conversion
    │       └── <ENTITY>.go
    ├── proxy/          # Proxy service: per–source-server listeners, route lookup, reverse proxy
    │   └── service.go
    └── ui_server/...   # HTTP server struct, NewServer(addr, repo), Routes(), Run(ctx)
```

## Design patterns

- **Layered flow**: `main` → `server` (HTTP) → `database.Repository` (interface) → `database` handler (GORM). HTTP and business logic depend on **schema** types and the **Repository** interface only; they do not import `database/objects` or GORM.
- **Schema vs objects**:  
  - **`internal/database/schema`**: Domain/API types. Used in handlers, repository interface, and any app/cache logic. No GORM; JSON tags for API.  
  - **`internal/database/objects`**: Persistence-only types. GORM struct tags, `TableName()`, soft delete, etc. Conversion lives here: `XToSchema` / `SchemaToX` so the rest of the app never touches ORM types.
- **Dependency injection**: `main` builds `Handler` → `Repository(db)` → `Server(addr, repo)`. Server holds `Repository` and uses it in handlers; no direct DB access in `internal/server`.
- **HTTP stack**: Standard library only. `http.ServeMux` in `Routes()`; no third-party router. API under `/api/`, UI and static assets at `/`, `/styles.css`, `/app.js`. Static files are embedded in `ui.go`.
- **Config**: Database driver and DSN come from `.env` (`DB_DRIVER`, `DB_DSN`). `main` loads env (e.g. godotenv) before creating the DB handler.

## Proxy

- **`internal/proxy`**: Runs one HTTP listener per source server (bind to each source’s host:port). For each request, looks up a route by (source server, method, path) via `Repository.FindRouteBySourceMethodPath`, loads the target server, builds the backend URL (target protocol, host, port, base_path, route target_path, query), and forwards the request with `net/http/httputil.ReverseProxy`. Route matching is **exact** (method + path). The package depends only on `database.Repository` and `database/schema`; no GORM or `database/objects`. `main` starts the proxy service in parallel with the UI server using the same `context` for graceful shutdown.

## Conventions to follow

1. **New domain entities**: Add a type in `database/schema`, a matching type and conversions in `database/objects`, register in `Handler.AutoMigrate`, extend `Repository` and `repository` impl, then add API routes and handlers in `server` using schema types only.
2. **New API surface**: Add routes in `routes.go`, implement handlers in `handlers.go` (or a new file in `server`). Handlers call `s.repo` and work with `schema.*`; never use `objects.*` or `*gorm.DB` in `internal/server`.
3. **Static assets**: Add files under `internal/server/static/`, embed in `ui.go`, and register path in `Routes()`.
4. **Package boundaries**: `internal/server` and `internal/proxy` may import `internal/database` and `internal/database/schema` only (and stdlib). `internal/database` may import `internal/database/schema` and `internal/database/objects`. `internal/database/objects` may import `internal/database/schema`; keep GORM and DB details inside `database` and `objects`.

Keeping these layers and the schema/objects split consistent will keep the codebase predictable for both humans and AI.
