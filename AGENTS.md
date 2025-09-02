# Repository Guidelines

## Project Structure & Module Organization
- `handlers/`: HTTP handlers (11 files: chat, emotion, user, character, etc.)
- `services/`: Core logic (20 files: chat, NSFW analyzer, emotion, memory, TTS, AI clients, smart routing)
- `routes/`: Route registration (`routes.go`) with 47 API endpoints
- `models/`: Data models with 7 database tables using JSONB for complex data
  - `db/`: Database models (User, Character, CharacterSpeechStyle, CharacterScene, ChatSession, Message, Relationship)
- `cmd/bun/`: Unified CLI tool for database migrations and management
- `middleware/`: Auth and cross‑cutting concerns
- `utils/`: Logging, errors, helpers, JWT
- `public/`: Static UI + Swagger UI entry
- `bin/`: Build artifacts

## Build, Test, and Development Commands
- `make install`: Sync deps; install `swag`.
- `make run`: Start server locally (`main.go`).
- `make build`: Compile to `bin/thewavess-ai-core`.
- `make test`: Run `go test -v ./...`.
- `make docs` / `make docs-serve`: Generate Swagger and serve with the app.
- `make db-setup` / `make migrate` / `make migrate-status` / `make migrate-down`: Database management via unified CLI tool
- `./tests/test-all.sh`: Run unified test suite (24 tests, 100% pass rate).
- `make docker-build` / `make docker-run`: Containerize and run.

## Coding Style & Naming Conventions
- Language: Go 1.23+; keep code `go fmt` clean before pushing.
- Packages: lowercase, no underscores; files `snake_case.go` by feature.
- Exported identifiers: `UpperCamelCase`; locals: `lowerCamelCase`.
- JSON tags: snake_case (e.g., `json:"should_use_grok"`).
- Keep functions small and focused; prefer constructors in `services/` for DI.

## Testing Guidelines
- Place unit tests as `*_test.go` next to code; prefer table‑driven tests.
- Run all tests with `make test`.
- Use `make test-api` for endpoint smoke tests.
- Cover handlers (happy/error paths) and service logic; mock external APIs.

## Commit & Pull Request Guidelines
- Follow Conventional Commits used here: `feat: ...`, `docs: ...`.
- PRs include purpose, scope, screenshots (UI/Swagger), and migration notes.
- Link issues; note API changes and rollout steps.
- Keep diffs focused; include docs/tests updates.

## Security & Configuration Tips
- Configure via `.env` (see `.env.example`): `OPENAI_API_KEY`, `GROK_API_KEY`, `DATABASE_URL`, etc.
- Never commit secrets; verify `.gitignore` coverage.
- NSFW routing is automatic via `services/smart_nsfw_analyzer.go` with 5-level classification
- Memory system integrated into relationships.emotion_data JSONB field
- Tags system integrated into character.metadata.tags field
- Smart engine routing automatically selects optimal AI service (OpenAI/Grok)
