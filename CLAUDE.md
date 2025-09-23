# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
**Production-Ready System**: Female-oriented AI chat backend with Go + Gin, dual-engine AI (OpenAI + Grok), intelligent NSFW classification.

**Reference**: See [AGENTS.md](./AGENTS.md) for repository structure and development workflows.
**Documentation**: Complete documentation index available at [DOCS_INDEX.md](./DOCS_INDEX.md).

## Core Architecture
- **Backend**: Go 1.24 + Gin + Bun ORM + PostgreSQL
- **Dual AI Engine**: OpenAI GPT-4o (L1-L3) + Grok AI (L4-L5) with intelligent routing
- **NSFW System**: Keyword-based content classifier with zero runtime cost (~3.5μs)
- **Caching Layer**: Ristretto high-performance in-memory cache with TinyLFU algorithm
- **Prompt System**: Inheritance-based builders (`BasePromptBuilder` → engine-specific)

## Key Components
- **Chat Service**: `services/chat_service.go:selectAIEngine()` - Dual engine routing with fallbacks
- **NSFW Classifier**: `services/keyword_classifier_*.go` - Keyword matching (L1-L5), zero runtime cost
- **Character System**: `services/{character_service,character_store,character_cache}.go` - Dynamic personality + caching
- **Cache Layer**: `services/{character_cache,cache_service}.go` - Ristretto-based high-performance caching
- **Handlers**: Clean APIs in `handlers/` - real database fields, no fake data
- **Prompt Architecture**: `services/prompt_{base,openai,grok}.go` - DRY inheritance pattern
- **AI Clients**: `services/{openai,grok}_client.go` - Official SDK integration
- **TTS Service**: `services/tts_service.go` - OpenAI voice synthesis

## Performance & Caching
**Ristretto Cache Implementation**:
- **Character Cache**: `services/character_cache.go` - Eliminates duplicate DB queries (3 queries → 1)
- **Relationship Cache**: `services/cache_service.go` - Converted from Redis to Ristretto
- **Performance**: 5986x improvement (21.2ms → 3.5μs) for character lookup operations (eliminates duplicate DB queries)
- **Configuration**: Character cache 1MB, Relationship cache 2MB, TinyLFU eviction, async writes with 10ms delay

**Cache Strategy**: Check cache first, fallback to DB, then cache result with async write completion delay

## Message Regeneration System
**Architecture**: `services/chat_service.go:RegenerateMessage()` - Direct AI generation without ProcessMessage
**Handler**: `handlers/chat.go:/regenerate` - Updates existing message instead of creating new ones
**Features**:
- Multiple regenerations supported without duplicates
- AI failure handling with placeholder messages
- Database consistency maintained for user/AI message pairs

## Essential Commands
```bash
# Quick Setup
make fresh-start         # Complete fresh installation
make dev                # Generate docs + start server
make run                # Direct execution without docs generation

# Development
make build docs test-all # Standard development workflow
make db-setup fixtures   # Database initialization

# Testing
./tests/test-all.sh     # Complete test suite (24/24 tests)
```

## Development Standards
- **Architecture**: Service → Handler → Model layers, real data only
- **Database**: Go migrations, JSONB for complex data, Bun ORM
- **AI Integration**: Dual engine (OpenAI L1-L3, Grok L4-L5), sticky sessions, fallback chain
- **Caching**: Ristretto for all in-memory caching, Redis connection preserved for future use
- **Code Quality**: Go 1.24+ built-ins, no fake data, DRY principles

## NSFW Classification System
Keyword-based content classifier with zero runtime cost:
- **L5-L4**: Explicit content → Grok | **L3-L1**: Safe/moderate → OpenAI
- **Zero cost**: No API calls, microsecond-level response time (~3.5μs)
- **Features**: Sticky sessions (5min), fallback chain, illegal content blocking
- **Maintenance**: Keywords managed in source code, no external dependencies

## Error Handling & Recovery
**AI Response Failures**:
- Create placeholder messages in database to maintain conversation continuity
- Allow regeneration attempts without database gaps
- Maintain user/AI message pairing consistency

**Cache Failures**:
- Graceful fallback to database queries
- Automatic cache warming on successful DB retrievals
- No service interruption on cache misses

## Configuration & Status
**Required `.env`**: Database credentials, `OPENAI_API_KEY`, `GROK_API_KEY`, model names
**Optional**: `OPENAI_SEED`, `OPENAI_LOGPROBS`, `TTS_API_KEY`, `REDIS_*` (connection only)
**Status**: 57/57 APIs complete, 7 tables, 24/24 tests pass, 1-3s response, 95%+ NSFW accuracy, Swagger at :8080/swagger

## Troubleshooting
**Cache Issues**: Check Ristretto metrics, verify 10ms async write delay
**Fixtures**: `make migrate-reset && make fixtures`
**Build**: `make docs` first for Swagger generation
**Dependencies**: `make clean && make install`
**Database**: Ensure PostgreSQL running, check connection in logs

## Commit Guidelines
Follow existing style with detailed descriptions and co-authoring format.

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.