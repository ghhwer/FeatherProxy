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
    │   ├── repository.go   # Repository type alias, NewRepository, NewCachedRepository
    │   ├── repo/           # Repository interface and ErrProtocolMismatch (no impl deps)
    │   ├── impl/           # Concrete repository (GORM + cache); cache_helpers.go has getCached/invalidate and keys
    │   ├── cache/          # Cache interface (Get, Set, Delete, DeleteByPrefix); strategies: none, memory, redis stub
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

1. **New domain entities**: Add a type in `database/schema`, a matching type and conversions in `database/objects`, register in `Handler.AutoMigrate`, extend `Repository` in `repo/repo.go` and implement in `impl/` (with cache), then add API routes and handlers in `server` using schema types only. See **New entities with cache** below.
2. **New API surface**: Add routes in `routes.go`, implement handlers in `handlers.go` (or a new file in `server`). Handlers call `s.repo` and work with `schema.*`; never use `objects.*` or `*gorm.DB` in `internal/server`.
3. **Static assets**: Add files under `internal/server/static/`, embed in `ui.go`, and register path in `Routes()`.
4. **Package boundaries**: `internal/server` and `internal/proxy` may import `internal/database` and `internal/database/schema` only (and stdlib). `internal/database` may import `internal/database/schema` and `internal/database/objects`. `internal/database/objects` may import `internal/database/schema`; keep GORM and DB details inside `database` and `objects`.

## New entities with cache

Caching is built into the repository implementation in `impl/`. When adding a new entity you only touch `repo/`, `impl/`, and (as usual) schema, objects, handler, and server. No separate cache wrapper.

1. **Schema and objects** (unchanged): Add `schema/<ENTITY>.go`, `objects/<ENTITY>.go` with conversions, register the object in `Handler.AutoMigrate`.
2. **Repository interface**: Extend `repo/repo.go` with the new methods (e.g. `GetFoo(id)`, `CreateFoo`, `UpdateFoo`, `DeleteFoo`, `ListFoos`).
3. **Cache keys in `impl/cache_helpers.go`**: Add key constants and key builder funcs for the entity, e.g. `keyPrefixFoo = "foo:"`, `keyFoo(id)`, `keyListFoos = "list:foos"`. Follow the existing pattern (single-item keys like `entity:id`, list key like `list:entities`).
4. **Implementation in `impl/`** (new file e.g. `impl/foo.go` or in an existing entity file):
   - **Getters** (including list): Use `getCached(r, key, func() (T, error) { ... DB logic, return schema value ... })`. Put the actual GORM/objects logic inside the delegate; the helper handles cache read, JSON marshal/unmarshal, and cache write on miss.
   - **Mutations** (Create/Update/Delete/Set*): Do the DB work, then `return r.invalidate(err, keys, prefixes)`. Pass the DB error as first argument; on success pass the exact keys to delete (e.g. the entity key and the list key) and any prefixes to clear (e.g. `keyPrefixRoute` when any route changes). For a simple entity: `r.invalidate(r.db.Create(&obj).Error, []string{keyFoo(id), keyListFoos}, nil)`.
   - **Do not cache** methods that return sensitive or decrypted data (e.g. plain tokens). Call the DB directly and do not wrap in `getCached`.
5. **Invalidation rules**: On Create/Update/Delete of an entity, invalidate that entity’s key(s) and its list key. If the entity is referenced by others (e.g. routes reference servers), invalidate any cache entries that depend on it (e.g. when an auth is updated, we invalidate `keyPrefixTargetAuthForRoute`). See existing impl files for examples.

Keeping these layers and the schema/objects split consistent will keep the codebase predictable for both humans and AI.
