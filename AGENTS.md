# Repository Guidelines

## Project Structure & Module Organization
- `handlers/`: HTTP entry points (chat, emotion, user, character).
- `services/`: Core logic for chat, routing, memory, NSFW checks, TTS, and AI clients.
- `routes/routes.go`: Registers ~47 API endpoints.
- `models/db/`: Bun ORM models; JSONB fields hold complex payloads.
- `middleware/` and `utils/`: Auth, logging, error handling, helpers, JWT tools.
- `cmd/bun/`: Migration CLI; run alongside Makefile targets.
- `public/` and `bin/`: Static UI/Swagger assets and build artifacts.
- Tests live next to implementation files as `*_test.go`; integration scripts sit in `tests/`.

## Build, Test, and Development Commands
- `make install`: Sync Go modules and install Swagger tooling.
- `make run`: Start the server from `main.go` with hot reload via Go.
- `make build`: Produce `bin/thewavess-ai-core` binary.
- `make test`: Execute `go test -v ./...` across modules.
- `make docs` / `make docs-serve`: Generate Swagger specs and serve them with the app.
- `make db-setup`, `make migrate`, `make migrate-status`, `make migrate-down`: Provision and manage the Bun-backed database.

## Coding Style & Naming Conventions
- Target Go 1.23+; run `go fmt ./...` before committing.
- Package names remain lowercase without underscores; file names use snake_case by feature.
- Exported identifiers use UpperCamelCase; locals use lowerCamelCase.
- JSON tags follow `snake_case`, e.g. ``json:"should_use_grok"``.
- Prefer small functions with clear boundaries; create constructors in `services/` for DI.

## Testing Guidelines
- Write table-driven unit tests in `*_test.go` files next to the code.
- Mock external APIs and network calls; avoid flaky integration tests.
- Fast suite: `make test`; full integration: `./tests/test-all.sh`.
- Structure assertions for both success and failure paths.

## Commit & Pull Request Guidelines
- Use Conventional Commits, e.g. `feat: add smart router`, `fix: tts timeout`.
- PRs should explain purpose, scope, and linked issues; add Swagger screenshots when endpoints change.
- Include migration notes and rollout steps when schema or infra changes.
- Keep diffs focused, update docs/tests alongside code, and call out API adjustments.

## Security & Configuration Tips
- Configure secrets via `.env` (see `.env.example` for `OPENAI_API_KEY`, `GROK_API_KEY`, `DATABASE_URL`, etc.).
- Never commit credentials; confirm `.gitignore` coverage before pushing.
- Avoid logging sensitive payloads; NSFW routing, memory, and tagging are driven through `services/` and JSONB fields.
