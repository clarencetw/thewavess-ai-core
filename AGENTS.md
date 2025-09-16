# Repository Guidelines

## Project Structure & Module Organization
- `handlers/`: HTTP handlers (chat, emotion, user, character, etc.).
- `services/`: Core logic (chat, NSFW analyzer, emotion, memory, TTS, AI clients, smart routing).
- `routes/`: Route registration (`routes.go`) exposing ~47 endpoints.
- `models/db/`: Database models (`User`, `Character`, `CharacterSpeechStyle`, `CharacterScene`, `ChatSession`, `Message`, `Relationship`). Complex fields use JSONB.
- `cmd/bun/`: CLI for migrations and DB management.
- `middleware/`: Auth and cross‑cutting concerns. `utils/`: logging, errors, helpers, JWT.
- `public/`: Static UI + Swagger entry. `bin/`: build artifacts. Tests are co‑located as `*_test.go`; integration in `./tests/test-all.sh`.

## Build, Test, and Development Commands
- `make install`: Sync deps; install `swag`.
- `make run`: Start the server (`main.go`).
- `make build`: Compile to `bin/thewavess-ai-core`.
- `make test`: Run unit tests (`go test -v ./...`).
- `make docs` / `make docs-serve`: Generate Swagger and serve with the app.
- DB: `make db-setup`, `make migrate`, `make migrate-status`, `make migrate-down` (via `cmd/bun`).
- Docker: `make docker-build`, `make docker-run`.

## Coding Style & Naming Conventions
- Go 1.23+; code must be `go fmt` clean.
- Packages: lowercase, no underscores. Files use `snake_case.go` by feature (e.g., `smart_router.go`).
- Names: Exported `UpperCamelCase`; locals `lowerCamelCase`.
- JSON: snake_case tags (e.g., `json:"should_use_grok"`).
- Design: Keep functions small; use constructors in `services/` for DI.

## Testing Guidelines
- Tests live next to code as `*_test.go`; prefer table‑driven tests.
- Mock external APIs; cover handlers (happy/error) and service logic.
- Run unit tests with `make test`. Full suite: `./tests/test-all.sh`.

## Commit & Pull Request Guidelines
- Commits follow Conventional Commits (e.g., `feat: add smart router`, `fix: tts timeout`).
- PRs include purpose, scope, linked issues, screenshots (UI/Swagger), and migration notes.
- Keep diffs focused; update docs/tests alongside code; call out API changes and rollout steps.

## Security & Configuration Tips
- Configure via `.env` (see `.env.example`): `OPENAI_API_KEY`, `GROK_API_KEY`, `DATABASE_URL`, etc.
- Never commit secrets; verify `.gitignore` coverage.
- NSFW routing, memory system, and tags integrate via services/JSONB; smart routing selects OpenAI/Grok automatically.
