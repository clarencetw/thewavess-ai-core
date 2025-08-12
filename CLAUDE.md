# Claude Memory

## Project Overview
Female-oriented AI chat backend built with Go + Gin framework, integrating OpenAI GPT-4o with 5-level NSFW content classification.

## Key Components
- **Chat Service**: Core logic in `services/chat_service.go:analyzeNSFWContent()`
- **OpenAI Client**: Character prompts in `services/openai_client.go:getCharacterSystemPrompt()`
- **Logging**: Structured logging via `utils/logger.go`
- **Error Handling**: API errors in `utils/errors.go`

## Characters
- **Lu Hanyuan (陸寒淵)**: Dominant CEO character
- **Shen Yanmo (沈言墨)**: Gentle doctor character

## NSFW System
5-level content classification:
- **Level 1-4**: OpenAI handles (including sexual organ descriptions)
- **Level 5**: Grok handles (explicit content)
- **Adult-only**: 18+ age restriction enforced

## Current Status
- **API Progress**: 22/118 endpoints implemented
- **Performance**: 1-3s response time, 95%+ accuracy
- **Environment**: `.env` with OPENAI_API_KEY required
- **Testing**: Web interface at http://localhost:8080

## Important Files
- `API_PROGRESS.md`: Development progress tracking
- `NSFW_GUIDE.md`: Detailed NSFW implementation guide
- `DEPLOYMENT.md`: Deployment and operations guide