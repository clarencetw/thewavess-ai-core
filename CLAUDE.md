# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
**Production-Ready System**: Female-oriented AI chat backend built with Go + Gin framework, featuring dual-engine AI architecture with intelligent 5-level NSFW content classification. Complete chat functionality with emotional intelligence and character interaction.

**Note**: For OpenAI-specific development guidelines, see [AGENTS.md](./AGENTS.md) which contains repository structure, coding standards, and development workflows optimized for AI-assisted development.

## Core Architecture
- **Backend**: Go 1.23 + Gin + Bun ORM
- **Database**: PostgreSQL (main data) + Redis (cache/sessions)
- **Dual AI Engine**: OpenAI GPT-4o (L1-L3) + Grok AI (L4-L5) with intelligent routing
- **NSFW System**: 18-rule weighted keyword classifier with 95%+ accuracy
- **Prompt System**: Inheritance-based prompt builders (`BasePromptBuilder` → engine-specific builders)
- **Deployment**: Docker Compose with health checks

## Key Components
- **Chat Service**: `services/chat_service.go:selectAIEngine()` - Dual AI engine routing with fallback mechanisms
- **NSFW Classifier**: `services/nsfw_classifier.go` - Intelligent weighted keyword analysis (18 rules, L1-L5)
- **Prompt Architecture**:
  - `services/prompt_base.go` - Base prompt builder with shared functionality
  - `services/prompt_openai.go` - OpenAI-specific prompts (L1-L3 safe content)
  - `services/prompt_grok.go` - Grok-specific prompts (L4-L5 creative content)
  - `services/prompt_mistral.go` - Mistral prompts (preserved but unused in dual-engine mode)
- **AI Clients**: `services/{openai,grok,mistral}_client.go` - Engine-specific API integrations
- **Character System**: `services/character_service.go` + `services/character_store.go` - Dynamic character personality management
- **JSON Sanitizer**: `utils/json_sanitize.go` - Robust AI response parsing with mixed format support

## Essential Commands

### Quick Setup
```bash
make fresh-start    # Complete fresh installation
make quick-setup    # Database + fixtures only
make dev           # Generate docs + start server
```

### Development
```bash
make build         # Compile application
make docs          # Generate Swagger documentation
make test-all      # Run all test suites (24 tests, 100% pass rate)
make check         # Health check for running services
```

### Database Management
```bash
make db-setup      # Initialize + migrate database
make fixtures      # Load test data (characters, users)
make migrate-reset # Reset database (with confirmation)
make create-migration NAME=migration_name  # Create new Go migration
```

### Testing
```bash
./tests/test-all.sh                    # Complete test suite
./tests/chat_api_validation.sh         # Chat API + dual-engine validation
./tests/test_mistral_integration.sh    # Mistral engine integration tests
make test-integration                  # Integration tests
```

## Development Standards

### Go Architecture Patterns
- **Service Layer**: Business logic in `services/` - each service handles one domain
- **Handler Layer**: HTTP endpoints in `handlers/` - thin controllers that delegate to services
- **Model Layer**: Database models in `models/db/` using Bun ORM annotations
- **Middleware Stack**: `middleware/` for cross-cutting concerns (auth, logging, CORS)
- **Error Handling**: Structured error responses via `utils/errors.go:APIError`

### Database Standards
- **Migrations**: Go files in `cmd/bun/migrations/`, NOT SQL files
- **UPSERT Operations**: Require UNIQUE constraints, not regular indexes
- **JSONB Usage**: Store complex data (emotions, metadata) in JSONB fields
- **Bun ORM**: Use model definitions over raw SQL: `bunDB.NewCreateTable().Model((*Model)(nil)).IfNotExists().Exec(ctx)`

### AI Integration Patterns
- **Dual Engine Architecture**: OpenAI (L1-L3 safe→moderate) + Grok (L4-L5 explicit content)
- **Intelligent Routing**: Enhanced keyword classifier with flexible pattern matching, sticky sessions (5min) to prevent engine switching
- **Prompt Inheritance**: `BasePromptBuilder` provides shared functionality, engine-specific builders extend for specialized needs
- **Fallback Mechanisms**: OpenAI content rejection → automatic Grok fallback, Mistral errors → Grok fallback
- **Context Management**: 6 recent messages @ 120 chars each, chat/novel mode support
- **JSON Response Handling**: Mixed format parser supports both structured JSON and "content + --- + metadata" formats

## NSFW Classification System
The system uses a sophisticated 18-rule weighted keyword classifier:

### Classification Levels
- **L5 (≥10 points)**: Explicit sexual acts → Grok (高潮、射精、口交、肛交)
- **L4 (≥6 points)**: Explicit body parts → Grok (陰莖、陰道、生殖器)
- **L3 (≥4 points)**: Nudity/porn contexts → OpenAI* (裸體、色情、床戲)
- **L2 (≥2 points)**: Body descriptions → OpenAI* (胸部、身材、性感)
- **L1 (<2 points)**: Safe content → OpenAI (安全對話)

*Originally designed for Mistral in three-engine architecture, now handled by OpenAI in dual-engine mode

### Intelligent Features
- **Context Suppression**: Medical/art/education contexts auto-downgrade levels
- **Sticky Sessions**: 5-minute engine consistency after L4+ trigger (improved from 3min)
- **Illegal Content Blocking**: Taiwan law compliance (underage, violence, non-consent)
- **Fallback Chain**: OpenAI rejection → Grok, ensures service availability

## Prompt Builder Architecture
The system uses an inheritance-based prompt builder pattern:

### Core Pattern
```go
BasePromptBuilder  // Shared functionality (context, character, NSFW levels)
├── OpenAIPromptBuilder    // Safe content (L1-L3) + fallback logic
├── MistralPromptBuilder   // Preserved for future use
└── GrokPromptBuilder      // Creative content (L4-L5) + artistic enhancements
```

### Key Methods
- `WithCharacter(character)` - Inject character personality and traits
- `WithContext(conversationContext)` - Add chat history (6 messages @ 120 chars)
- `WithNSFWLevel(level)` - Set content safety level (L1-L5)
- `WithChatMode(mode)` - Switch between "chat" (conversational) and "novel" (narrative)
- `Build()` - Generate final prompt string for AI engine

### Chat vs Novel Modes
- **Chat Mode**: Natural conversation flow, concise responses (150-300 chars)
- **Novel Mode**: Rich narrative descriptions, detailed scenes (~2x content length)

## Environment Configuration
Required variables in `.env`:
- **Database**: `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- **AI APIs**: `OPENAI_API_KEY`, `GROK_API_KEY` (Mistral preserved: `MISTRAL_API_KEY`)
- **AI Models**: `OPENAI_MODEL=gpt-4o-mini`, `GROK_MODEL=grok-3-mini`
- **Security**: `JWT_SECRET`
- **CORS**: `CORS_ALLOWED_ORIGINS`

## Current System Status
- **API**: 47/47 endpoints implemented (100% complete)
- **Database**: 7 optimized tables, production-ready architecture
- **Performance**: 1-3s response time, 95%+ NSFW accuracy
- **Testing**: 24/24 tests passing (100% success rate)
- **Documentation**: Swagger UI at http://localhost:8080/swagger/index.html

## Troubleshooting
- **Fixtures Error**: Run `make migrate-reset` then `make fixtures`
- **Dependency Issues**: `make clean && make install`
- **AI Engine Failures**: System automatically falls back between OpenAI and Grok
- **Build Issues**: Check `docs/` generation with `make docs` before `make build`

## Testing Strategy
- **Integration Tests**: Full API workflow testing in `tests/`
- **Chat Validation**: Specialized NSFW and conversation flow testing
- **Health Checks**: Automated service monitoring via `make check`
- **Test Configuration**: Environment-specific settings in `tests/test-config.sh`

## Commit Guidelines
Follow existing commit style with detailed descriptions and co-authoring format.