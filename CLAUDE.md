# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Thewavess AI Core is a Golang-based intelligent chat backend that integrates multiple AI engines to provide smart conversation capabilities for Thewavess products.

### AI Engine Strategy
- **Primary Chat**: OpenAI GPT-4o for core language communication
- **NSFW Content**: Grok for adult/sensitive content handling
- **Text-to-Speech**: OpenAI TTS service as default
- **Vector Database**: Qdrant for semantic memory search

## Current State

This repository is in its initial state with only a README.md file present. The project appears to be a fresh Golang project that needs to be set up from scratch.

## Development Setup

Since this is a new repository, development commands and build processes have not yet been established. When working on this project:

1. Initialize Go module: `go mod init github.com/clarencetw/thewavess-ai-core`
2. Set up basic project structure for a Go backend service
3. Add necessary dependencies for AI engine integrations
4. Implement HTTP server and API endpoints
5. Add configuration management for different AI providers
6. Implement proper logging and error handling

## Architecture Notes

Key architectural considerations for this AI chat backend:

- **AI Engine Router**: Intelligent routing between OpenAI (default) and Grok (NSFW) based on content analysis
- **Content Classification**: NSFW detection system to determine appropriate AI engine
- **Multi-Modal Support**: Text chat with GPT-4o and TTS audio generation with OpenAI TTS
- **RESTful API**: Standard HTTP endpoints for chat, TTS, and engine management
- **Configuration Management**: Environment-based config for API keys and engine settings
- **Service Layer**: Abstraction layer for different AI providers (OpenAI, Grok)
- **Middleware Stack**: Authentication, logging, rate limiting, and content filtering