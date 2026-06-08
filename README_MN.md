# Gerege Template AI

> 🌐 [English](README.md) · **Монгол**

[![Go](https://img.shields.io/badge/Go-1.26-blue.svg)](https://golang.org/)
[![Fiber](https://img.shields.io/badge/Fiber-v3-00ACD7.svg)](https://gofiber.io/)
[![Next.js](https://img.shields.io/badge/Next.js-14-black.svg)](https://nextjs.org/)
[![Claude](https://img.shields.io/badge/AI-Anthropic%20Claude-d97757.svg)](https://www.anthropic.com/)
[![Gemini](https://img.shields.io/badge/Voice-Google%20Gemini-4285F4.svg)](https://ai.google.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Org](https://img.shields.io/badge/Org-geregedevops-181717.svg?logo=github)](https://github.com/geregedevops)
[![Repo](https://img.shields.io/badge/GitHub-geregedevops%2Fgovernment--template--platform-181717.svg?logo=github)](https://github.com/geregedevops/government-template-platform)

Clean Architecture зарчмаар бүтээгдсэн, аюулгүй байдлыг хатууруулсан,
production-д бэлэн **AI-native full-stack template**. Go (**Fiber v3 + GORM +
PostgreSQL + Redis**) backend болон Next.js (**BFF**) frontend-ийг хослуулж, хоёр
AI гадаргууг end-to-end холбосон:

- **AI чат туслах** — **Anthropic Claude**-д суурилсан streaming (SSE) чат,
  ярианы түүх, өөрийгөө тайлбарлах system prompt-той. Чат нь **дуу хоолой**-г
  бас дэмжинэ: асуултаа ярьж асууж (STT), хариуг нь сонсож (TTS) болно.
- **Дуу хоолойн орчуулга** — **Google Gemini**-д суурилсан **Монгол ↔ Англи**
  яриа орчуулга: дуу→бичвэр → орчуулга → бичвэр→дуу, тоглуулахад бэлэн аудиогоор
  буцаана.

Хоёр AI үйлчилгээний API түлхүүр зөвхөн **backend талд** байна — browser руу
хэзээ ч дамжихгүй (BFF загвар).

## 📌 Эх сурвалж ба нээлттэй эх

**Backend** нь нээлттэй эх
[snykk/go-rest-boilerplate](https://github.com/snykk/go-rest-boilerplate)
(MIT, Najib Fikri)-аас гаралтай; HTTP давхаргыг **Gin → Fiber v3**, өгөгдлийн
давхаргыг **sqlx → GORM** болгож хөрвүүлсэн, бүх фичерийг хадгалсан. Fiber v3-ийн
хэрэглээг [rachmanzz/fiber-starter](https://github.com/rachmanzz/fiber-starter)
(MIT)-ээс лавласан. Бүх эх төслийн зохиогчийн эрх, лицензийг хүндэтгэн хадгалсан
([LICENSE](LICENSE), [NOTICE](NOTICE), [AUTHORS](AUTHORS)). Энэ төсөл **MIT
лицензтэй**.

## Monorepo бүтэц

```
gerege-template-ai/
├── backend/    # Go · Fiber v3 · GORM · PostgreSQL · Redis · JWT/OTP танилт
│   ├── internal/business/usecases/ai/      # Claude чатын usecase
│   ├── internal/business/usecases/voice/   # Gemini STT/орчуулга/TTS usecase
│   ├── pkg/aiclient/                        # Anthropic Messages (SSE) клиент
│   ├── pkg/geminiclient/                    # Gemini generateContent клиент
│   └── docs/   # ARCHITECTURE · DEVELOPMENT · API_CONTRACT · SECURITY (EN/MN)
└── frontend/   # Next.js BFF — /chat (AI) ба /translate (дуу хоолой) хуудас
```

- **[backend/README_MN.md](backend/README_MN.md)** — Clean Architecture Go API.
- **[frontend/README.md](frontend/README.md)** — Next.js Backend-for-Frontend.

## Онцлог

- **Clean Architecture** — `handler → usecase → repository → domain`, back-import байхгүй; business core нь web framework-ийг import хийдэггүй.
- **AI чат (Anthropic Claude)** — streaming SSE хариу, яриа + мессежийн түүх (RLS-ээр хязгаарлагдсан), хэрэглэгчийн өдрийн лимит, хоёр хэлний system prompt. `/api/v1/ai/*` endpoint-ууд.
- **Дуу хоолой (Google Gemini)** — MN↔Англи яриа орчуулга, чатын STT/TTS. Аудио оролт/гаралт base64-аар; backend нь Gemini-ийн PCM-г WAV болгон боодог. `/api/v1/voice/*` endpoint-ууд.
- **Танилт** — JWT access + refresh (rotation), OTP-баталгаажуулсан бүртгэл, bcrypt, login lockout.
- **Аюулгүй хатууруулсан** — security headers (per-request nonce-той CSP, HSTS, Permissions-Policy), CORS allow-list, rate limiting, request timeout, parameterized query, Postgres Row-Level Security. [SECURITY.md](SECURITY.md)-г үз.
- **Observability** — OpenTelemetry trace + Prometheus metrics + Zap structured log.
- **Frontend BFF** — браузер зөвхөн ижил-origin Next.js route рүү залгаж, тэр нь server талаас backend руу проксиолдог (токен **болон** AI түлхүүр client JS-д хүрэхгүй).
- **Тесттэй** — unit + testcontainers integration тест.

## Түргэн эхлүүлэх

**Шаардлага:** Go 1.26+, Node 20+, PostgreSQL 15+, Redis 7+. Сонголтоор:
Anthropic + Gemini API түлхүүр (тохируулаагүй бол AI/voice endpoint-ууд `503` буцаана).

### 0) Go module-ийн нэрийг солих (нэг удаа)

Template нь `geregetemplateai` module замаар ирдэг. Өргөтгөхөөсөө өмнө өөрийн
module зам болгон солино уу — бүх .go файл `geregetemplateai/...`-г import хийдэг.

```bash
./scripts/rename-module.sh github.com/myorg/my-api
cd backend && go mod tidy && cd ..
```

### 1) Backend ба Frontend

```bash
# 1) Backend  →  http://localhost:8080
cd backend
cp internal/config/.env.example internal/config/.env   # JWT_SECRET (≥32), DB, Redis тохируул
# Сонголтоор AI: ANTHROPIC_API_KEY (чат), GEMINI_API_KEY (дуу хоолой) тохируул
make mig-up                                            # схем (AI + voice хүснэгт орно)
make serve

# 2) Frontend →  http://localhost:3000
cd ../frontend
cp .env.example .env.local                              # BACKEND_URL=http://localhost:8080
npm install
npm run dev
```

**http://localhost:3000** нээж, бүртгүүлэх / нэвтэрсний дараа **/chat** (AI) болон
**/translate** (дуу хоолой)-г туршина уу. Дуу хоолойд browser-ийн микрофон
зөвшөөрөл + HTTPS эсвэл `localhost` origin шаардлагатай.

## Баримтжуулалт

| Doc | Юу |
|-----|------|
| [backend/docs/ARCHITECTURE_MN.md](backend/docs/ARCHITECTURE_MN.md) | Давхаргууд, dependency flow (AI + voice орсон) |
| [backend/docs/DEVELOPMENT_MN.md](backend/docs/DEVELOPMENT_MN.md) | Фичер нэмэх заавар, тест, code style |
| [backend/docs/API_CONTRACT_MN.md](backend/docs/API_CONTRACT_MN.md) | REST endpoint, request/response (`/ai/*`, `/voice/*` орсон) |
| [backend/docs/SECURITY.md](backend/docs/SECURITY.md) | Хэрэгжсэн хяналт + ASVS roadmap |
| [SECURITY.md](SECURITY.md) | Эмзэг байдлыг хэрхэн мэдээлэх |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Хэрхэн хувь нэмэр оруулах |

## Хувь нэмэр

Хувь нэмэр оруулахыг урьж байна — [CONTRIBUTING.md](CONTRIBUTING.md) болон
[Code of Conduct](CODE_OF_CONDUCT.md)-ийг уншина уу.

## Лиценз

[MIT](LICENSE) — snykk/go-rest-boilerplate (MIT)-ийн derivative. Эх төслийн
мэдэгдлийг [LICENSE](LICENSE), [NOTICE](NOTICE)-д хадгалсан.

---

**Gerege Template AI v1.0** — **Gerege Systems Development Team** болон
**Claude AI** хамтран бүтээв, 2026.
