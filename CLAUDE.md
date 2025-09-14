# Claude Code Instructions

## Project Overview
**Production-Ready System**: Female-oriented AI chat backend built with Go + Gin framework, integrating OpenAI GPT-4o with 5-level NSFW content classification. Complete chat functionality with emotional intelligence and character interaction.

**Note**: For OpenAI-specific development guidelines, see [AGENTS.md](./AGENTS.md) which contains repository structure, coding standards, and development workflows optimized for AI-assisted development.

## Core Architecture
- **Backend**: Go 1.23 + Gin + Bun ORM
- **Database**: PostgreSQL (main data) + Redis (cache/sessions)
- **AI Services**: OpenAI GPT-4o + Grok API
- **NSFW System**: 5-level classification (L1-4: OpenAI, L5: Grok, Adult 18+)
- **Deployment**: Docker Compose with health checks

## Key Components
- **Chat Service**: `services/chat_service.go:analyzeNSFWContent()`
- **OpenAI Client**: `services/openai_client.go:getCharacterSystemPrompt()`
- **Database**: `database/` directory with Bun ORM migrations
- **Error Handling**: `utils/errors.go` for consistent API responses

## Development Standards

### Go Best Practices
- Use Bun ORM model definitions over raw SQL
- Follow error handling patterns in `utils/errors.go`
- Implement structured logging via `utils/logger.go`
- Never expose secrets or API keys in logs/commits

### Frontend Standards
- **CSS Framework**: Tailwind CSS utilities only
- **Components**: Flowbite for modals, notifications, complex UI
- **Mobile First**: Ensure responsive design across all pages
- **No Custom CSS**: Use utility classes, preserve only essential animations

### Database Standards
- **Migrations**: Go files in `cmd/bun/migrations/`, NOT SQL files
- **UPSERT Operations**: Require UNIQUE constraints, not regular indexes
- **JSONB Usage**: Store complex data (emotions, metadata) in JSONB fields

## Essential Commands

### Quick Setup
```bash
make fresh-start    # Complete fresh installation
make quick-setup    # Database + fixtures only
make dev           # Generate docs + start server
```

### Database Management
```bash
make db-setup      # Initialize + migrate database
make fixtures      # Load test data (characters, users)
make migrate-reset # Reset database (with confirmation)
```

### Development
```bash
make build         # Compile application
make test-all      # Run all test suites
make docs          # Generate Swagger documentation
```

### Troubleshooting
- **Fixtures Error**: Run `make migrate-reset` then `make fixtures`
- **Dependency Issues**: `make clean && make install`

## Environment Configuration
Required variables in `.env`:
- **Database**: `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- **AI APIs**: `OPENAI_API_KEY`, `GROK_API_KEY`
- **Security**: `JWT_SECRET`
- **CORS**: `CORS_ALLOWED_ORIGINS`

## Current System Status
- **API**: 47/47 endpoints implemented (100% complete)
- **Database**: 7 optimized tables, production-ready architecture
- **Performance**: 1-3s response time, 95%+ NSFW accuracy
- **Testing**: 24/24 tests passing (100% success rate)
- **Documentation**: Swagger UI at http://localhost:8080/swagger/index.html

## Technical Notes

### Database Migration System
- All migrations in `cmd/bun/migrations/` as Go files
- Create migrations: `go run cmd/bun/main.go create-migration <name>`
- Use Bun ORM model definitions: `bunDB.NewCreateTable().Model((*Model)(nil)).IfNotExists().Exec(ctx)`

### NSFW Content Classification
5-level AI-powered intelligent classification system:
- **Levels 1-4**: OpenAI GPT-4o with GPT-5-nano classification
- **Level 5**: Grok API (explicit content)
- **Classification Method**: Pure AI analysis via GPT-5-nano ($0.05/1M tokens)
- **Age Restriction**: 18+ enforced for adult content

## Commit Guidelines
Follow existing commit style with detailed descriptions and co-authoring format.