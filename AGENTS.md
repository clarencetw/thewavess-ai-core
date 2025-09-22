# Repository Guidelines

> 📋 **Related Documentation**: Complete documentation index available at [DOCS_INDEX.md](./DOCS_INDEX.md)

## Project Structure & Module Organization
- `handlers/` hosts HTTP entry points for chat, emotion, user, and character flows.
- Core business logic lives in `services/`, including routing, memory, NSFW checks, TTS, and AI client adapters.
- API wiring is centralized in `routes/routes.go` (~57 endpoints); Bun models sit in `models/db/` with JSONB payload fields.
- Middleware, utilities, and JWT helpers reside in `middleware/` and `utils/`; migrations run via `cmd/bun/`. Tests accompany source files as `*_test.go`, with broader scripts in `tests/`.

## Build, Test, and Development Commands
- `make install` syncs modules and Swagger tooling.
- `make run` starts `main.go` with hot reload; use during iterative development.
- `make build` produces the `bin/thewavess-ai-core` binary for deployment.
- `make test` executes `go test -v ./...` across all packages.
- `make docs` regenerates Swagger specs.
- Database lifecycle: `make db-setup`, `make migrate`, `make migrate-status`, `make migrate-down`.

## Coding Style & Naming Conventions
- Target Go 1.23+; always run `go fmt ./...` before commits.
- Packages use lowercase names; files adopt snake_case by feature.
- Exported identifiers follow UpperCamelCase; locals use lowerCamelCase. JSON tags stay snake_case (e.g. `json:"should_use_grok"`).
- Keep functions focused and prefer constructors in `services/` for dependency injection.

## Testing Guidelines
- Write table-driven unit tests alongside implementations (`*_test.go`).
- Cover both success and failure paths; mock external APIs to avoid flaky runs.
- Quick suite: `make test`. Full integration: `./tests/test-all.sh`.

## Commit & Pull Request Guidelines
- Use Conventional Commits (e.g. `feat: add smart router`, `fix: tts timeout`).
- PRs should explain scope, purpose, linked issues, and include Swagger screenshots when API endpoints change.
- Document migrations and rollout steps whenever database schema or infra shifts.

## Security & Configuration Tips
- Configure secrets via `.env`; see `.env.example` for expected keys.
- Never log sensitive payloads; NSFW handling, memory, and tagging logic run through `services/` and JSONB-backed models.
- Ensure `.gitignore` excludes credentials before pushing.
