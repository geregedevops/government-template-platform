# Gerege Template AI

> 🌐 **English** · [Монгол](README_MN.md)

[![Go](https://img.shields.io/badge/Go-1.26-blue.svg)](https://golang.org/)
[![Fiber](https://img.shields.io/badge/Fiber-v3-00ACD7.svg)](https://gofiber.io/)
[![Next.js](https://img.shields.io/badge/Next.js-14-black.svg)](https://nextjs.org/)
[![Claude](https://img.shields.io/badge/AI-Anthropic%20Claude-d97757.svg)](https://www.anthropic.com/)
[![Gemini](https://img.shields.io/badge/Voice-Google%20Gemini-4285F4.svg)](https://ai.google.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)
[![Org](https://img.shields.io/badge/Org-geregedevops-181717.svg?logo=github)](https://github.com/geregedevops)
[![Repo](https://img.shields.io/badge/GitHub-geregedevops%2Fgovernment--template--platform-181717.svg?logo=github)](https://github.com/geregedevops/government-template-platform)

A production-ready, security-hardened **AI-native full-stack template** built on
Clean Architecture. It pairs a Go (**Fiber v3 + GORM + PostgreSQL + Redis**)
backend with a Next.js (**BFF**) frontend, and ships with two AI surfaces wired
end-to-end:

- **AI chat assistant** — streaming (SSE) chat powered by **Anthropic Claude**,
  with per-conversation history and a self-describing system prompt. The chat
  also supports **voice**: speak your question (STT) and listen to the reply (TTS).
- **Voice translation** — **Mongolian ↔ English** speech translation powered by
  **Google Gemini**: speech-to-text → translate → text-to-speech, returned as
  playable audio.

Both AI providers' API keys live **only on the backend** — they never reach the
browser (BFF pattern).

## 📌 Origin & Open Source

The **backend** is derived from the open-source
[snykk/go-rest-boilerplate](https://github.com/snykk/go-rest-boilerplate)
(MIT, by Najib Fikri); we ported the HTTP layer **Gin → Fiber v3** and the data
layer **sqlx → GORM**, keeping the full feature set. Fiber v3 idioms were
cross-referenced against [rachmanzz/fiber-starter](https://github.com/rachmanzz/fiber-starter)
(MIT). All upstream copyright and license terms are honored — see [LICENSE](LICENSE),
[NOTICE](NOTICE), and [AUTHORS](AUTHORS). This project is **MIT-licensed**.

## Monorepo structure

```
gerege-template-ai/
├── backend/    # Go · Fiber v3 · GORM · PostgreSQL · Redis · JWT/OTP auth
│   ├── internal/business/usecases/ai/      # Claude chat usecase
│   ├── internal/business/usecases/voice/   # Gemini STT/translate/TTS usecase
│   ├── pkg/aiclient/                        # Anthropic Messages (SSE) client
│   ├── pkg/geminiclient/                    # Gemini generateContent client
│   └── docs/   # ARCHITECTURE · DEVELOPMENT · API_CONTRACT · SECURITY (EN/MN)
└── frontend/   # Next.js BFF — /chat (AI) and /translate (voice) pages
```

- **[backend/README.md](backend/README.md)** — Clean Architecture Go API.
- **[frontend/README.md](frontend/README.md)** — Next.js Backend-for-Frontend.

## Features

- **Clean Architecture** — `handler → usecase → repository → domain`, no back-imports; the business core never imports the web framework.
- **AI chat (Anthropic Claude)** — streaming SSE responses, conversation + message history (RLS-scoped), per-user daily limit, bilingual system prompt. Endpoints under `/api/v1/ai/*`.
- **Voice (Google Gemini)** — MN↔English voice translation plus chat STT/TTS. Audio in/out as base64; backend wraps Gemini PCM in WAV. Endpoints under `/api/v1/voice/*`.
- **Auth** — JWT access + refresh (rotation), OTP-verified registration, bcrypt, login lockout.
- **Security-hardened** — strict security headers (CSP with per-request nonce, HSTS, Permissions-Policy), CORS allow-list, rate limiting, request timeouts, parameterized queries, Postgres Row-Level Security. See [SECURITY.md](SECURITY.md).
- **Observability** — OpenTelemetry tracing + Prometheus metrics + structured Zap logs.
- **Frontend BFF** — the browser talks only to same-origin Next.js routes, which proxy to the backend server-side (tokens **and** AI keys never reach client JS).
- **Tested** — unit tests + testcontainers integration tests.

## Quick start

**Prerequisites:** Go 1.26+, Node 20+, PostgreSQL 15+, Redis 7+. Optional:
Anthropic + Gemini API keys (the AI/voice endpoints return `503` until set).

### 0) Rename the Go module (one-time)

The template ships under the module path `geregetemplateai`. Replace it with your
own module path before extending — every Go file imports `geregetemplateai/...`,
so renaming early avoids a sed sweep later.

```bash
./scripts/rename-module.sh github.com/myorg/my-api
cd backend && go mod tidy && cd ..
```

### 1) Backend → http://localhost:8080

```bash
cd backend
cp internal/config/.env.example internal/config/.env   # set JWT_SECRET (≥32), DB, Redis
# Optional AI: set ANTHROPIC_API_KEY (chat) and GEMINI_API_KEY (voice)
make mig-up        # create schema (includes AI + voice tables)
make serve
```

### 2) Frontend → http://localhost:3000

```bash
cd ../frontend
cp .env.example .env.local                              # BACKEND_URL=http://localhost:8080
npm install
npm run dev
```

Open **http://localhost:3000**, register / log in, then try **/chat** (AI) and
**/translate** (voice). Voice needs a browser mic permission and an HTTPS or
`localhost` origin.

### 3) Docker (full stack)

```bash
cp backend.env.example backend.env     # fill secrets incl. ANTHROPIC_API_KEY / GEMINI_API_KEY
# create ./.env with POSTGRES_*, REDIS_PASS, APP_DB_*, APP_ORIGIN, WEB_PORT
docker compose up -d --build
```

## Documentation

| Doc | What |
|-----|------|
| [backend/docs/ARCHITECTURE.md](backend/docs/ARCHITECTURE.md) | Layers, dependency flow, components (incl. AI + voice) |
| [backend/docs/DEVELOPMENT.md](backend/docs/DEVELOPMENT.md) | Add-a-feature guide, testing, code style |
| [backend/docs/API_CONTRACT.md](backend/docs/API_CONTRACT.md) | REST endpoints, request/response shapes (incl. `/ai/*`, `/voice/*`) |
| [backend/docs/SECURITY.md](backend/docs/SECURITY.md) | Implemented controls + ASVS roadmap |
| [SECURITY.md](SECURITY.md) | How to report a vulnerability |
| [CONTRIBUTING.md](CONTRIBUTING.md) | How to contribute |

## Contributing

Contributions are welcome — please read [CONTRIBUTING.md](CONTRIBUTING.md) and
the [Code of Conduct](CODE_OF_CONDUCT.md).

## License

[MIT](LICENSE) — derivative of snykk/go-rest-boilerplate (MIT). Upstream notices
are retained in [LICENSE](LICENSE) and [NOTICE](NOTICE).

---

**Gerege Template AI v1.0** — Co-developed by the **Gerege Systems
Development Team** and **Claude AI**, 2026.
