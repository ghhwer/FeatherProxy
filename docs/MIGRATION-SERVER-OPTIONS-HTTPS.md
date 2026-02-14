# Migration: Server Options & HTTPS (Protocol-Specific Implementations)

This document plans the migration from a single HTTP-only proxy implementation to **protocol-aware source servers** with a new **Server Options** entity, enabling HTTPS (with server-side TLS) and keeping the design abstract for future protocols (e.g. gRPC, TCP).

---

## 1. Goals

- **HTTP vs HTTPS**: Route traffic to different implementations. For HTTPS, the **server side** (FeatherProxy listener) must present a certificate; options (e.g. cert/key) live in a dedicated entity.
- **Server Options entity**: New entity related to **Source Server** (1:1), holding protocol-specific options (e.g. TLS cert path, key path for HTTPS). Extensible for future protocols.
- **Abstract server implementation**: Proxy layer chooses listener type by protocol (+ options), so we can add new protocols later without changing the core flow.

---

## 2. Current State

| Layer | Current behavior |
|-------|-------------------|
| **Schema** | `SourceServer` has `Protocol` (string, e.g. `http`, `https`), `Host`, `Port`, `Name`. No options. |
| **Proxy** | One `http.Server` per source; always `ListenAndServe()` (no TLS). Protocol is stored but not used to change listener type. |
| **UI** | Create/Edit source: name, protocol (http/https), host, port. No certificate or options. |

---

## 3. Data Model: Server Options Entity

### 3.1 Relationship

- **ServerOptions** ↔ **SourceServer**: **1:1** (one options row per source server).
- Optional: a source server can have no options (e.g. HTTP needs none). For HTTPS we require options with TLS data.

### 3.2 Schema (domain)

**`database/schema/server_options.go`**

- `SourceServerUUID uuid.UUID` — FK to source server (unique).
- `TLSCertPath string` — path to server certificate file (PEM). Empty for non-HTTPS.
- `TLSKeyPath string` — path to private key file (PEM). Empty for non-HTTPS.
- Optional later: `TLSCertPEM`, `TLSKeyPEM` (for DB-stored certs; consider encryption at rest).
- Optional: generic `Options map[string]string` or `OptionsJSON []byte` for future protocol options (e.g. gRPC settings) without schema changes.

Recommendation: start with **TLS path-based** only; add PEM-in-DB later if needed.

### 3.3 Objects (ORM)

**`database/objects/server_options.go`**

- Table name: `server_options`.
- Columns: `source_server_uuid` (PK, FK to `source_servers.source_server_uuid`), `tls_cert_path`, `tls_key_path`, `created_at`, `updated_at`, `deleted_at` (soft delete).
- Conversions: `ServerOptionsToSchema`, `SchemaToServerOptions`.

### 3.4 Repository

**`repo/repo.go`**

- `GetServerOptions(sourceServerUUID uuid.UUID) (schema.ServerOptions, error)` — return options or zero value if not found.
- `SetServerOptions(opts schema.ServerOptions) error` — upsert by `SourceServerUUID`.
- Optional: `DeleteServerOptions(sourceServerUUID uuid.UUID) error` — when source server is deleted, options are deleted (or cascade).

**`impl/`**

- Implement with cache: cache key e.g. `server_options:{source_server_uuid}`. Invalidate on Set/Delete and when source server is deleted.

### 3.5 Validation rules

- For **HTTPS** source servers: either require ServerOptions with non-empty `TLSCertPath` and `TLSKeyPath`, or allow creation but fail at proxy start with a clear error. Recommendation: allow saving; validate at proxy start and log/skip or fail.
- For **HTTP**: ignore options for listener; options may still be stored (e.g. future use or pre-config for protocol switch).

---

## 4. Database Migration

- **AutoMigrate**: Register `&objects.ServerOptions{}` in `Handler.AutoMigrate()` (GORM will create or alter table).
- No separate migration files in current setup; if you introduce versioned migrations later, add a step that creates `server_options` and, if needed, backfills or leaves nulls.

---

## 5. Proxy Layer: Abstract Server Implementation

### 5.1 Strategy

- **Per-source runner**: For each source server, the proxy decides **how** to listen based on `source.Protocol` and, for HTTPS, `server_options`.
- **Abstraction**: Introduce a small **listener factory** or **runner** abstraction so adding a new protocol (e.g. gRPC) is a new implementation, not a big branch in one function.

### 5.2 Proposed abstraction

**Option A – Interface per protocol**

```go
// ListenerRunner starts a listener for one source server and blocks until ctx is done.
type ListenerRunner interface {
    Run(ctx context.Context, source schema.SourceServer, options schema.ServerOptions) error
}
```

- `HTTPRunner`: ignores options, uses `http.Server` + `ListenAndServe()`.
- `HTTPSRunner`: uses `TLSCertPath` and `TLSKeyPath` from options, `ListenAndServeTLS(certFile, keyFile)`.
- Proxy service: `sources := repo.ListSourceServers()`; for each source, `opts := repo.GetServerOptions(source.UUID)`; select runner by `source.Protocol`; `go runner.Run(ctx, source, opts)`.

**Option B – Single service, internal strategy**

- Keep one `Service.Run(ctx)`; inside the loop, switch on `source.Protocol`: `case "https": runHTTPS(ctx, source, opts)`, `default: runHTTP(ctx, source)`. Less abstract but simpler; refactor to Option A when adding a third protocol.

Recommendation: **Option B** for the first iteration (HTTP + HTTPS only); extract to **Option A** when adding a second protocol (e.g. gRPC).

### 5.3 Concrete behavior

- **HTTP**: unchanged — `http.Server` + `ListenAndServe()`.
- **HTTPS**:
  - Load options: `opts, err := repo.GetServerOptions(source.SourceServerUUID)`.
  - If `opts.TLSCertPath == "" || opts.TLSKeyPath == ""`: log error and skip this source (or fail entire proxy start), e.g. `log.Printf("proxy: HTTPS source %s missing TLS cert/key paths, skipping", source.Name)`.
  - Else: `server.ListenAndServeTLS(opts.TLSCertPath, opts.TLSKeyPath)` (same handler as HTTP).

Handler logic (route lookup, reverse proxy, auth) stays shared; only the way the listener is started changes.

### 5.4 Repository dependency

- Proxy already depends on `database.Repository`. Add usage of `GetServerOptions(sourceServerUUID)` when starting each source. No new dependencies.

---

## 6. API & Handlers

### 6.1 Options as sub-resource of source server

- **GET** `/api/source-servers/{uuid}/options` — return ServerOptions for that source (404 if none).
- **PUT** `/api/source-servers/{uuid}/options` — body: `{ "tls_cert_path": "...", "tls_key_path": "..." }`; upsert.
- Optional: **DELETE** `/api/source-servers/{uuid}/options` — remove options row.

Alternatively, **embed options in source server response**:

- **GET** `/api/source-servers/{uuid}` — include `server_options: { tls_cert_path, tls_key_path }` when present.
- **PUT** `/api/source-servers/{uuid}` — accept optional `server_options` in body and upsert in one shot.

Recommendation: **sub-resource** (`/options`) for clearer separation and to avoid sending cert paths on list. List source servers stays lightweight; options loaded only when editing or when proxy starts.

### 6.2 Handlers

- New handler functions in `handlers/`: `GetServerOptions`, `SetServerOptions` (and optionally `DeleteServerOptions`).
- Reuse existing UUID parsing and JSON response helpers; validate that the source server exists before get/set options.

---

## 7. UI Changes

### 7.1 Source server create/edit

- **Protocol** remains (http / https). When user selects **HTTPS**:
  - Show extra fields: **TLS certificate path**, **TLS key path** (text inputs or file pickers that resolve to paths).
  - On submit (create or update): after creating/updating the source server, call **PUT** `/api/source-servers/{uuid}/options` with `{ tls_cert_path, tls_key_path }`.
- When **HTTP** is selected: hide cert/key fields; optionally **PUT** options with empty paths or leave options unchanged.

### 7.2 Load options when editing

- When opening **Edit** for a source server: **GET** `/api/source-servers/{uuid}/options` and fill cert/key fields if present. If 404, treat as empty.

### 7.3 List view

- Optional: show an indicator (e.g. lock icon or “TLS”) for source servers that have HTTPS and non-empty options. No need to show cert paths in the table.

### 7.4 Validation and errors

- Client-side: if HTTPS and cert or key path empty, show warning (e.g. “TLS paths required for HTTPS; proxy may skip this server until set”).
- After save, “Reload proxies” so the proxy process restarts and picks up new options.

---

## 8. Implementation Order (Checklist)

1. **Schema & objects**
   - [ ] Add `schema/server_options.go` (SourceServerUUID, TLSCertPath, TLSKeyPath).
   - [ ] Add `objects/server_options.go` (table, conversions).
   - [ ] Register `ServerOptions` in `Handler.AutoMigrate()`.

2. **Repository**
   - [ ] Add `GetServerOptions`, `SetServerOptions` to `repo/repo.go`.
   - [ ] Add cache keys in `impl/cache_helpers.go` (e.g. `server_options:{uuid}`).
   - [ ] Implement in `impl/server_options.go` (getCached, invalidate on set; invalidate on source server delete if cascade).

3. **API & handlers**
   - [ ] Add routes: `GET/PUT /api/source-servers/{id}/options` in `routes.go`.
   - [ ] Add `handlers.GetServerOptions`, `handlers.SetServerOptions` (and optional Delete).

4. **Proxy**
   - [ ] In `proxy/service.go`: for each source, load `GetServerOptions(source.SourceServerUUID)`.
   - [ ] Branch on `source.Protocol`: `"https"` → `ListenAndServeTLS(opts.TLSCertPath, opts.TLSKeyPath)`; else → `ListenAndServe()`.
   - [ ] When HTTPS and missing cert/key: log and skip (or fail) that source.

5. **UI – API client**
   - [ ] In `api.js`: add `getSourceServerOptions(uuid)`, `setSourceServerOptions(uuid, body)`.

6. **UI – Forms**
   - [ ] In create/edit source modals: when protocol is HTTPS, show TLS cert path and TLS key path inputs.
   - [ ] On submit (create): after create source, call `setSourceServerOptions(uuid, { tls_cert_path, tls_key_path })`.
   - [ ] On submit (edit): same; load options in `editSource(uuid)` and fill form.

7. **Docs & reload**
   - [ ] Update `AGENTS.md` or README if you document the new entity and options behavior.
   - [ ] Ensure “Reload proxies” still restarts the proxy so new options are applied.

---

## 9. Future Extensions (Abstract Server Implementation)

- **More protocols**: Add e.g. `grpc` to `Protocol`. Implement `GRPCRunner` that implements `ListenerRunner` and register it in a map by protocol. ServerOptions can gain `options_json` or new columns for gRPC (e.g. TLS, reflection).
- **PEM in DB**: Add `tls_cert_pem`, `tls_key_pem` (or encrypted) to ServerOptions; in HTTPS runner, if paths are empty but PEM is set, use `tls.X509KeyPair(certPEM, keyPEM)` and `TLSConfig.Certificates`. Requires secure handling of secrets (e.g. encryption at rest, no logging).
- **Optional TLS for HTTP**: Keep HTTP as non-TLS; no change. If you ever want “HTTP with optional TLS”, the same options entity can be reused with a different protocol name or flag.

---

## 10. Summary

| Area | Change |
|------|--------|
| **New entity** | ServerOptions (1:1 with SourceServer): TLS cert path, key path; extensible later. |
| **DB** | New table `server_options`; AutoMigrate; repo get/set with cache. |
| **API** | GET/PUT `/api/source-servers/{uuid}/options`. |
| **Proxy** | Branch on `source.Protocol`: HTTPS → `ListenAndServeTLS` with options; HTTP → `ListenAndServe`. Abstract to a runner interface when adding more protocols. |
| **UI** | Create/Edit source: when HTTPS, show and save TLS paths; load options when editing. |

This keeps the **server implementation abstract** (protocol-driven listener choice) and **expandable** (new protocols = new runner + optional new options fields) while delivering HTTPS with server-side certificates via the new **Server Options** entity tied to **Source Server**.
