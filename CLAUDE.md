# Claude Code Instructions

## Project Overview
Female-oriented AI chat backend built with Go + Gin framework, integrating OpenAI GPT-4o with 5-level NSFW content classification.

## Key Components
- **Chat Service**: Core logic in `services/chat_service.go:analyzeNSFWContent()`
- **OpenAI Client**: Character prompts in `services/openai_client.go:getCharacterSystemPrompt()`
- **Database**: Connection and migrations in `database/connection.go`
- **Logging**: Structured logging via `utils/logger.go`
- **Error Handling**: API errors in `utils/errors.go`

## Technical Preferences
- **CSS Framework**: Tailwind CSS - minimize custom CSS, use utility classes
- **UI Components**: Flowbite for notifications, modals, and complex components
- **Responsive Design**: Mobile-first approach with responsive breakpoints
- **Code Style**: Pure Tailwind classes preferred over custom CSS
- **Environment**: Docker Compose with comprehensive .env support

## NSFW System
5-level content classification:
- **Level 1-4**: OpenAI GPT-4o handles (including sexual descriptions)
- **Level 5**: Grok handles (explicit content)
- **Adult-only**: 18+ age restriction enforced
- **Classification**: Automated keyword detection with 95%+ accuracy

## Architecture
- **Backend**: Go 1.23 + Gin + Bun ORM
- **Database**: PostgreSQL (main data) + Redis (cache/sessions)
- **Vector DB**: Qdrant (memory search, optional)
- **AI Services**: OpenAI GPT-4o + Grok API
- **Deployment**: Docker Compose with health checks

## Environment Configuration
All environment variables support .env file and docker-compose.yml:
- **Database**: `DB_HOST`, `DB_USER`, `DB_PASSWORD`, etc.
- **AI APIs**: `OPENAI_API_KEY`, `GROK_API_KEY`, model configurations
- **Security**: `JWT_SECRET`
- **CORS**: `CORS_ALLOWED_ORIGINS`, headers, methods

## Development Best Practices
- **Mobile Responsiveness**: Ensure all pages work well on phones
- **Notification System**: Use Flowbite Toast, avoid custom CSS notifications
- **CSS Guidelines**: 
  - Preserve only essential custom CSS (line-clamp, animations)
  - Replace custom styles with Tailwind utilities when possible
  - Maintain consistent spacing and typography
- **Docker**: Use multi-stage builds, include all necessary directories

## CLI Commands
Unified CLI tool commands following Bun best practices:

### Database Management
- **Initialize**: `make db-init` - Initialize migration tables
- **Migrate**: `make migrate` - Run pending migrations  
- **Rollback**: `make migrate-down` - Rollback last migration
- **Status**: `make migrate-status` - Show migration status
- **Reset**: `make migrate-reset` - Reset all migrations (with confirmation)
- **Setup**: `make db-setup` - Complete database setup (init + migrate)

### Data Seeding
- **Seed**: `make seed` - Populate character seed data
- **Reset Seed**: `make seed-reset` - Reset and re-populate seed data
- **Preview**: `make seed-dry` - Dry run preview of seed operation

### Development
- **Install**: `make install` - Install dependencies
- **Build**: `make build` - Compile application
- **Run**: `make run` - Start development server with logging
- **Dev Mode**: `make dev` - Generate docs + start server
- **Clean**: `make clean` - Clean build artifacts

### Documentation
- **Generate**: `make docs` - Generate Swagger documentation
- **Serve**: `make docs-serve` - Generate docs and start server

### Testing & Health
- **Test**: `make test` - Run tests
- **API Test**: `make test-api` - Test API endpoints
- **Health Check**: `make check` - Check service status

### Quick Workflows
- **Fresh Start**: `make fresh-start` - Clean + install + db-setup + seed
- **Quick Setup**: `make quick-setup` - Database setup + seed only

### Direct CLI Tool Usage
All commands use the unified CLI tool: `go run cmd/bun/main.go <command>`
- Database operations: `go run cmd/bun/main.go db <subcommand>`
- Create migrations: `go run cmd/bun/main.go create_sql <name>`

## Current Status
- **API Progress**: 22/118 endpoints implemented
- **Performance**: 1-3s response time, 95%+ NSFW accuracy
- **Docker**: Go 1.23, full env variables, web interface included
- **Testing**: Web interface at http://localhost:8080
- **Documentation**: Swagger UI at http://localhost:8080/swagger/index.html

## Documentation Files

### Core Specifications
- `SPEC.md`: Complete product specification and roadmap
- `API.md`: API documentation and endpoints overview
- `README.md`: Project overview and quick start guide

### Development & Progress
- `API_PROGRESS.md`: Development progress tracking (22/118 endpoints)
- `CLI_MIGRATION_GUIDE.md`: CLI tool migration guide and command mapping

### System Design Guides
- `CHARACTER_GUIDE.md`: Character system design and configuration
- `MEMORY_GUIDE.md`: Memory system design (short-term/long-term)
- `EMOTION_GUIDE.md`: Emotion system implementation guide
- `AFFECTION_GUIDE.md`: Affection system and relationship mechanics
- `NSFW_GUIDE.md`: NSFW content classification system (5-level)

### Operations & Deployment
- `DEPLOYMENT.md`: Deployment and operations guide
- `MONITORING_GUIDE.md`: System monitoring and health check strategies

### Development Tools
- `AGENTS.md`: AI agent configurations and behaviors
- `.claude/agents/`: Specialized agent configurations for different development tasks

## Important Technical Notes

### Database Migration System
- **Migration Location**: All migrations are in `cmd/bun/migrations/` as Go files, NOT in `database/migrations/`
- **Creating Migrations**: Use `go run cmd/bun/main.go create-migration <name>` to create new Go-based migrations
- **Reset Database**: Use `make migrate-reset` for complete database reset (requires confirmation)
- **ON CONFLICT Requirements**: PostgreSQL `ON CONFLICT` syntax requires UNIQUE constraints, not regular indexes
- **Memory System Fix**: Added `unique_user_character_memory` constraint to `long_term_memories` table for proper UPSERT operations

### Important Technical Notes
- **Migration System**: Uses Bun Go-based migrations in `cmd/bun/migrations/`, NOT SQL files
- **PostgreSQL UPSERT**: `ON CONFLICT` operations require UNIQUE constraints, not regular indexes
- **Database Schema**: All table constraints must be properly defined for UPSERT operations

## Commit Guidelines
Follow existing commit style with detailed descriptions and co-authoring format.