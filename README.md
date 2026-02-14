# FeatherProxy
FeatherProxy is an OSS proxy gateway for your API routing needs.

## Architecture

```mermaid
flowchart LR
  subgraph clients [Clients]
    C[Client]
  end
  subgraph app [FeatherProxy app]
    UI[UI server :4545]
    Mgr[ProxyManager]
    subgraph listeners [Per-source listeners]
      L1[Listener A host:port]
      L2[Listener B host:port]
    end
    Repo[Repository]
    subgraph persistence [Persistence]
      DB[(Database)]
      Cache[(Cache optional)]
    end
  end
  subgraph backends [Target backends]
    T1[Target 1]
    T2[Target 2]
  end
  C --> L1
  C --> L2
  L1 --> Mgr
  L2 --> Mgr
  Mgr --> T1
  Mgr --> T2
  UI --> Repo
  Mgr --> Repo
  Repo --> DB
  Repo -.-> Cache
```

- **Repository** — Single persistence interface for source/target servers, routes, and authentications. Used by both the UI server and the proxy.
- **Caching** — Optional: `CACHING_STRATEGY` (`none`, `memory`, or `redis`) and `CACHE_TTL` in env. When enabled, the repository caches reads (e.g. routes, server lists, auth metadata) and invalidates on writes. Sensitive data (e.g. decrypted tokens) is never cached.
- **Authentication** — Stored in the repository with tokens encrypted at rest. The UI uses the repository for CRUD (masked tokens in responses). The proxy uses `GetTargetAuthenticationWithPlainToken` (DB only, not cached) when forwarding requests to backends.