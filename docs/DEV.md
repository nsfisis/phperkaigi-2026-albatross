# Developer Guide

## Architecture

The system consists of the following components:

- **Reverse proxy (Nginx)** — Production only. Routes `/api/` and `/admin/` to the API server, serves the frontend SPA with fallback routing, and serves uploaded files from `/files/`.
- **API server (Go/Echo)** — Backend REST API and admin panel. Runs on port 8004 in development.
- **Frontend (React/Vite)** — Single-page application. In development, Vite dev server runs on port 5173 and proxies API requests.
- **Database (PostgreSQL 16.3)** — Primary data store.
- **Task queue (Redis 7.4 / Asynq)** — Asynchronous job processing for code submissions.
- **PHP worker (Node.js/Hono + WASM)** — Executes PHP submissions in a WebAssembly-sandboxed PHP interpreter.
- **Swift worker (Go/Echo + SwiftWasm/Wasmtime)** — Compiles and executes Swift submissions via SwiftWasm and Wasmtime.

Docker Compose orchestrates 6 services in development: `api-server`, `db`, `task-db`, `worker-php`, `worker-swift`, and `tools` (on-demand utilities like psqldef and sqlc).

## Project Structure

```
├── backend/            Go API server (Echo)
│   ├── schema.sql      Database schema
│   ├── query.sql       SQL queries (sqlc input)
│   ├── fixtures/       Dev seed data
│   ├── db/             Generated DB code (sqlc output)
│   ├── gen/            Generated API code (oapi-codegen output)
│   ├── admin/          Admin panel (server-rendered HTML)
│   ├── api/            REST API handlers
│   ├── game/           Game logic and WebSocket hub
│   ├── tournament/     Tournament bracket logic
│   └── justfile        Backend-specific commands
├── frontend/           React 19 / Vite 7 SPA
│   ├── app/            Application source
│   │   ├── api/        Generated OpenAPI types
│   │   ├── components/ React components
│   │   ├── hooks/      Custom hooks
│   │   ├── pages/      Page components
│   │   └── states/     Jotai state atoms
│   └── justfile        Frontend-specific commands
├── worker/
│   ├── php/            PHP WASM execution engine (Node.js/Hono)
│   │   └── justfile
│   └── swift/          Swift WASM execution engine (Go/Echo)
│       └── justfile
├── typespec/           TypeSpec API specifications
│   ├── api-server/     Main API spec
│   └── fortee/         Worker API spec
├── openapi/            Generated OpenAPI YAML files
├── docs/               Developer documentation
├── compose.local.yaml  Docker Compose for development
├── compose.prod.yaml   Docker Compose for production
├── justfile            Root build commands
└── package.json        npm workspaces root (frontend, typespec, worker/php)
```

## Dependencies

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- [Node.js](https://nodejs.org/) 22 or later (with npm)
- [Go](https://go.dev/) 1.25 or later
- [`just`](https://github.com/casey/just) command runner
- [direnv](https://direnv.net/) (optional, for auto-loading `.env`)

## Getting Started

1. Clone the repository.
2. `cd path/to/the/repo`
3. `cp .env.example .env`
4. `direnv allow .` (optional — loads env vars automatically)
5. `just init` — Builds Docker images, runs `npm install`, applies the database schema, and loads dev fixtures.
6. `just up` — Starts all Docker services and the Vite dev server.
7. Open http://localhost:5173/phperkaigi/2026/code-battle/.

**Test accounts:**

| Username | Password | Role |
|---|---|---|
| `a` | `pass` | Player |
| `b` | `pass` | Player |
| `c` | `pass` | Administrator |

## Development Commands

All commands are run from the repository root using `just`.

| Command | Description |
|---|---|
| `just` | Default: runs `down`, `build`, then `up` |
| `just build` | Build Docker images and run `npm install` |
| `just up` | Start Docker services and Vite dev server |
| `just down` | Stop all Docker services |
| `just logs` | View Docker service logs |
| `just logsf` | Follow Docker service logs |
| `just init` | Full initialization (`build` + `initdb`) |
| `just initdb` | Apply DB schema and load dev fixtures |
| `just gen` | Run the full code generation pipeline |
| `just check` | Run all linters/checks across all modules |
| `just test` | Run all tests across all modules |
| `just psql` | Open an interactive PostgreSQL shell |
| `just psql-query` | Non-interactive PostgreSQL (for piping SQL) |
| `just sqldef` | Apply schema migrations with psqldef |
| `just sqldef-dryrun` | Dry-run schema migrations |
| `just asynq` | Open the Asynq task queue dashboard |

Each module (`backend/`, `frontend/`, `worker/php/`, `worker/swift/`) also has its own `justfile` with `check`, `test`, and `ci` targets.

## Code Generation

The `just gen` command runs the full generation pipeline:

1. **TypeSpec to OpenAPI** — `npm -w typespec run build`
   - Input: `typespec/api-server/` and `typespec/fortee/`
   - Output: `openapi/api-server.yaml` and `openapi/fortee.yaml`
2. **Backend code generation** — `cd backend; go generate ./...`
   - Uses oapi-codegen to generate Go types and server stubs from the OpenAPI specs.
   - Also runs sqlc to generate type-safe Go code from `backend/query.sql`.
3. **Frontend type generation** — `npm -w frontend run openapi-typescript`
   - Input: `openapi/api-server.yaml`
   - Output: `frontend/app/api/schema.d.ts`

Run `just gen` after modifying TypeSpec definitions or SQL queries.

## Testing and Linting

### Running all tests

```
just test
```

This runs tests across all modules:

- **Backend**: `go test ./...`
- **Frontend**: `vitest run`
- **Worker PHP**: `vitest run`
- **Worker Swift**: `go test ./...`

### Running all checks

```
just check
```

This runs linters and type checks across all modules:

- **Backend**: `go build -o /dev/null ./...` + `golangci-lint run`
- **Frontend**: Biome (`biome check --write`) + TypeScript (`tsc --noEmit`) + ESLint
- **Worker PHP**: Biome (`biome check --write`)
- **Worker Swift**: `go build -o /dev/null ./...` + `golangci-lint run`

Each module's `just ci` target combines both check and test, and is what runs in CI.

## Database

- **Engine**: PostgreSQL 16.3, managed via Docker.
- **Schema**: `backend/schema.sql`
- **Dev fixtures**: `backend/fixtures/dev.sql`
- **Migrations**: Applied with [psqldef](https://github.com/sqldef/sqldef) via `just sqldef`. Use `just sqldef-dryrun` to preview changes.
- **Query management**: Queries are defined in `backend/query.sql` and [sqlc](https://sqlc.dev/) generates type-safe Go code in `backend/db/`.
- **Interactive shell**: `just psql` opens a `psql` session connected to the dev database.

## CI

GitHub Actions (`.github/workflows/ci.yml`) runs on pushes to `main` and on pull requests:

| Job | Runtime | Command |
|---|---|---|
| Backend | Go 1.25 | `just ci` (build + lint + test) |
| Frontend | Node.js 22 | `npm ci -w frontend` then `just ci` (vitest + biome + tsc + eslint) |
| Worker PHP | Node.js 22 | `npm ci -w worker/php` then `just ci` (biome + vitest) |
| Worker Swift | Go 1.25 | `just ci` (build + lint + test) |

## Environment Variables

- Copy `.env.example` to `.env` before running. It contains:
  - `ALBATROSS_BASE_PATH` — Base URL path for the application (default: `/phperkaigi/2026/code-battle/`).
- The `.envrc` file uses `dotenv_if_exists` so direnv will auto-load `.env` if present.
- Docker Compose internally sets database credentials (`ALBATROSS_DB_*`) and `ALBATROSS_IS_LOCAL=1` for the API server. These do not need manual configuration.
