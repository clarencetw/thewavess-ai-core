# Repository Guidelines

This guide orients contributors and automation agents to the codebase and workflow. Keep changes focused, well‑tested, and aligned with the patterns below.

## Project Structure & Module Organization
- `handlers/`: HTTP handlers (chat, emotion, user, character, etc.).
- `services/`: Core logic (chat, NSFW analyzer, emotion, memory, TTS, AI clients, smart routing).
- `routes/`: Route registration (`routes.go`) exposing ~47 endpoints.
- `models/db/`: Bun models (`User`, `Character`, `Message`, etc.); complex fields use JSONB.
- `cmd/bun/`: CLI for migrations and DB management.
- `middleware/`, `utils/`: Auth, logging, errors, helpers, JWT.
- `public/`: Static UI + Swagger entry. `bin/`: build artifacts.
- Tests live next to code as `*_test.go`; integration in `./tests/test-all.sh`.

## Build, Test, and Development Commands
- `make install`: Sync Go deps and install `swag`.
- `make run`: Start server (`main.go`).
- `make build`: Compile to `bin/thewavess-ai-core`.
- `make test`: Run unit tests (`go test -v ./...`).
- `make docs` / `make docs-serve`: Generate Swagger and serve with the app.
- DB: `make db-setup`, `make migrate`, `make migrate-status`, `make migrate-down`.
- Docker: `make docker-build`, `make docker-run`.

## Coding Style & Naming Conventions
- Go 1.23+; code must be `go fmt` clean.
- Packages: lowercase, no underscores. Files use snake_case by feature (e.g., `smart_router.go`).
- Names: Exported `UpperCamelCase`; locals `lowerCamelCase`.
- JSON tags: snake_case (e.g., ``json:"should_use_grok"``).
- Design: Small functions, clear boundaries; prefer constructors in `services/` for DI.

## Testing Guidelines
- Prefer table‑driven tests; mock external APIs and network calls.
- Cover handlers (happy/error paths) and service logic.
- Run fast suite with `make test`; full integration via `./tests/test-all.sh`.
- Keep tests next to implementation files as `*_test.go`.

## Commit & Pull Request Guidelines
- Commits follow Conventional Commits (e.g., `feat: add smart router`, `fix: tts timeout`).
- PRs include purpose, scope, linked issues, screenshots (UI/Swagger), and migration notes.
- Keep diffs focused; update docs/tests with code; call out API changes and rollout steps.

## Security & Configuration Tips
- Configure via `.env` (see `.env.example`): `OPENAI_API_KEY`, `GROK_API_KEY`, `DATABASE_URL`, etc.
- Never commit secrets; verify `.gitignore` coverage and avoid logging sensitive data.
- NSFW routing, memory, and tags integrate via services/JSONB; smart routing selects OpenAI/Grok automatically.

