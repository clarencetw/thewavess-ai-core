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

## Current Status
- **API Progress**: 22/118 endpoints implemented
- **Performance**: 1-3s response time, 95%+ NSFW accuracy
- **Docker**: Go 1.23, full env variables, web interface included
- **Testing**: Web interface at http://localhost:8080
- **Documentation**: Swagger UI at http://localhost:8080/swagger/index.html

## Important Files
- `SPEC.md`: Complete product specification and roadmap
- `API_PROGRESS.md`: Development progress tracking
- `NSFW_GUIDE.md`: Detailed NSFW implementation guide
- `DEPLOYMENT.md`: Deployment and operations guide

## Commit Guidelines
Follow existing commit style with detailed descriptions and co-authoring format.