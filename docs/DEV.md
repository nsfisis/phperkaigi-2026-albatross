# Architecture

* Reverse proxy server (Nginx)
* API server (Golang/Echo)
* Frontend (TypeScript/React/Vite)
* Database (PostgreSQL)
* Worker (Golang/Echo + Swift/SwiftWasm + WebAssembly/Wasmtime)

# Dependencies

* Docker
* Docker Compose
* Node.js 20.0.0 or later
* Npm
* Go 1.22.3 or later

# Run

1. Clone the repository.
1. `cd path/to/the/repo`
1. Copy `.env.example` to `.env` and configure:
    * `ALBATROSS_JWT_SECRET`: Secret key for JWT tokens
    * `ALBATROSS_COOKIE_SECRET`: Secret key for cookies
1. `direnv allow .` (optional)
1. `just init`
1. `just up`
1. Access to http://localhost:5173/iosdc-japan/2025/code-battle/.
    * User `a`, `b` and `c` can log in with `pass` password.
    * User `a` and `b` are players.
    * User `c` is an administrator.
