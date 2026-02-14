# FeatherProxy — Root Layout & Makefile (for AI)

This file describes the **repository root** and what the **Makefile** does so AI assistants can run and build the project correctly.

## Root layout

```
FeatherProxy/
├── Makefile       # run / build from app/
├── app/           # Go application (see app/AGENTS.md for internal structure)
├── README.md
├── LICENSE
└── .gitignore
```

All application code lives under `app/`. The root Makefile only delegates into `app/`; it does not define Go modules or source paths at the repo root.

## Makefile

- **Variables**
  - `APP_DIR := app` — Directory containing the Go app and `main.go`.
  - `BINARY := featherproxy` — Name of the compiled binary produced by `build`.

- **`make run`**
  - Runs the app in development: `cd app && go run main.go`.
  - Use this to start the server (e.g. http://localhost:4545). Ensure `app/.env` exists with `DB_DRIVER` and `DB_DSN` (see `app/.env.example`).

- **`make build`**
  - Builds the binary: `cd app && go build -o featherproxy .`
  - Output is `app/featherproxy`. The Makefile prints `Binary: app/featherproxy` after a successful build.
  - Run the binary from the repo root or from `app/`; it will use `app/.env` if you run from `app/`, or you must set env/config as needed.

- **`.PHONY: run build`**
  - Declares `run` and `build` as phony targets so `make` always runs their recipes even if files named `run` or `build` exist.

## Summary

| Target  | Action |
|---------|--------|
| `make run`   | `cd app && go run main.go` — run app in dev |
| `make build` | `cd app && go build -o featherproxy .` — produce `app/featherproxy` |

For the app’s internal design (database, server, schema/objects), see **`app/AGENTS.md`**.
