# Gerege AI Template Platform — Архитектурын санал ба код review

Огноо: 2026-06-04
Зорилго: Одоогийн Gerege template-ийг **өөрийгөө удирдах (self-managing), өөрийгөө тайлбарлах (self-explaining)**, бүрэн AI-суурьтай, Монгол/Англи хоёр хэлтэй платформ болгох.

AI хуваарилалт:

| Үүрэг | AI |
|---|---|
| STT, TTS, live voice translation | **Gemini** (Live API, native audio) |
| Chat, agent, RAG, self-explain, кодын туслах, орчуулгын pipeline | **Claude** (Opus 4.6 / Sonnet 4.6 / Haiku 4.5) |

---

## 1. Одоогийн кодын review — дүгнэлт

### Backend (Go) — Үнэлгээ: A

**Давуу талууд:**

- Clean Architecture — давхаргууд цэвэр тусгаарлагдсан (`cmd/`, `internal/business`, `internal/http`, `internal/datasources`, `pkg/`), import cycle байхгүй
- Аюулгүй байдал маш сайн: JWT algorithm pinning (HS256), access/refresh `Kind` тусгаарлалт, token cutoff (нууц үг солиход хуучин token хүчингүй болдог), bcrypt cost 12, timing attack хамгаалалт (dummy hash)
- PostgreSQL **RLS бүрэн идэвхтэй** (FORCE RLS, `SET LOCAL app.user_id/app.user_role`), default-deny
- Rate limiting (IP-ээр, /auth дээр 5 req/min), CORS, security headers, body size limit, request timeout
- OpenTelemetry + Prometheus, graceful shutdown, distroless Docker, non-root
- Typed domain error (`apperror`), testcontainers integration test

**Сул талууд:**

1. **API алдааны мессеж зөвхөн англиар** — `Accept-Language` дэмждэггүй, i18n library байхгүй
2. HTTP handler давхаргын test coverage сул (тест 37/150 файл)
3. `pkg/audit` байгаа ч бүрэн холбогдоогүй — нууц үг солих, role өөрчлөх үйлдэл audit log-гүй
4. Rate limit зөвхөн IP-ээр — хэрэглэгч тус бүрийн quota алга (AI endpoint-д заавал хэрэгтэй болно)
5. CSRF хамгаалалт байхгүй (SPA-д OK, гэхдээ баримтжуулах хэрэгтэй)

### Frontend (Next.js 14, BFF) — Үнэлгээ: A (i18n: D)

**Давуу талууд:**

- BFF архитектур: token-ууд **httpOnly + SameSite=Strict cookie**-д, browser JS-д ил гардаггүй
- Per-request CSP nonce, SSRF guard, refresh race-condition lock, idempotent logout
- TypeScript strict, `any` огт байхгүй, accessibility (ARIA, keyboard nav) сайн
- OKLCH design system, hydration mismatch-аас сэргийлсэн (Asia/Ulaanbaatar hardcode)
- Хамаарал маш цөөн (next, react, lucide-react л байна)

**Сул талууд:**

1. **i18n систем огт байхгүй** — Монгол текст компонент дотор hardcode, англи хувилбар ~30%
2. `error.tsx` boundary алга — server component алдаа гарвал хуудас унана
3. Request ID backend руу дамждаггүй — trace хийх боломжгүй
4. Frontend тест огт алга
5. Reset-password token URL-аас авдаг (referrer leak эрсдэл)

### Хоёр талд нийтлэг

- **AI код одоогоор 0 мөр** — "with-ai" нэртэй ч AI давхарга бүхэлдээ шинээр баригдана
- i18n хоёр талдаа дутуу — AI платформ болохоос ӨМНӨ эхэлж засах №1 зүйл

---

## 2. Зорилтот архитектур

```
┌──────────────────────────── Browser ────────────────────────────┐
│  Next.js UI (mn/en)   AI Chat UI   Voice UI (Mic/Speaker)        │
└──────┬──────────────────────┬───────────────┬────────────────────┘
       │ HTTPS (BFF)          │ SSE stream    │ WebSocket (audio)
┌──────▼──────────────────────▼───────────────▼────────────────────┐
│              Next.js BFF (cookie auth, CSP, proxy)                │
└──────┬──────────────────────┬───────────────┬────────────────────┘
       │                      │               │
┌──────▼──────────┐  ┌────────▼─────────────────────────┐
│  Go Backend     │  │  AI Gateway (Go, мөн backend дотор │
│  (одоогийнх:    │  │  internal/business/ai модуль)      │
│  auth, users,   │  │  • Claude API (chat, agent, RAG)   │
│  RLS, audit)    │  │  • Gemini Live proxy (STT/TTS/     │
│                 │  │    voice translation)              │
│                 │  │  • Tool registry + audit           │
│                 │  │  • Token/cost metering             │
└──────┬──────────┘  └────────┬──────────────┬───────────┘
       │                      │              │
┌──────▼──────────────────────▼──┐   ┌───────▼─────────┐
│ PostgreSQL (+pgvector, RLS)    │   │ Anthropic API    │
│ Redis (session, rate, cache)   │   │ Gemini Live API  │
└────────────────────────────────┘   └─────────────────┘
```

**Шийдвэр: тусдаа microservice бус, Go backend дотор `internal/business/ai` модуль.** Одоогийн clean architecture, RLS, auth, observability-г шууд дахин ашиглана. Хэрэгцээ гарвал дараа нь салгахад хялбар.

### 2.1 Модель сонголт (Claude)

| Ажил | Модель | Шалтгаан |
|---|---|---|
| Платформын чат, self-explain Q&A | `claude-sonnet-4-6` | Чанар/үнийн тэнцвэр |
| Routing, орчуулгын богино даалгавар, classification | `claude-haiku-4-5-20251001` | Хурдан, хямд |
| Self-managing agent, код үүсгэх, нарийн дүн шинжилгээ | `claude-opus-4-6` | Хамгийн чадалтай |

### 2.2 Gemini — дуу хоолой

- **STT + TTS + live translation**: Gemini Live API (native audio, bidirectional WebSocket)
- Урсгал: Browser mic → BFF WebSocket → AI Gateway → Gemini Live → audio/text буцаана
- **API түлхүүрийг browser-т ХЭЗЭЭ Ч өгөхгүй** — Gemini Live-ийн ephemeral token ашиглах, эсвэл бүх audio-г gateway-ээр дамжуулах (proxy). Эхний хувилбарт proxy-г санал болгож байна: audit, rate limit, кост хяналт нэг цэгт төвлөрнө
- Voice translation горим: mn→en, en→mn live орчуулга (хурал, харилцагчийн үйлчилгээ)

---

## 3. "Өөрийгөө тайлбарлах" (Self-Explaining) давхарга

Платформ өөрийн архитектур, API, тохиргоо, төлвөө хэрэглэгчид тайлбарлаж чаддаг байх.

1. **RAG knowledge base**: `docs/*.md` (ARCHITECTURE, API_CONTRACT, DEVELOPMENT, SECURITY — MN/EN хоёулаа), `swagger.json`, migration файлууд, README-үүдийг pgvector-т embed хийнэ. CI дээр docs өөрчлөгдөх бүрт дахин index хийнэ
2. **Платформын туслах чат** (`/ai/assistant`): "Энэ платформ RLS-ийг яаж хэрэгжүүлсэн бэ?", "Шинэ endpoint яаж нэмэх вэ?" гэх асуултад өөрийн кодын баримтаас иш татаж хариулна
3. **Live төлвийн тайлбар**: tool-ууд — `get_health`, `get_metrics`, `get_recent_errors`, `get_migrations_status`. Claude эдгээрийг дуудаж "Яагаад удаан байна?" гэх асуултад бодит metric-ээр хариулна
4. **API self-documentation**: swagger.json-г tool болгож өгснөөр "login хийх curl жишээ өгөөч" гэхэд үнэн зөв хариулна
5. Хариулт бүр хэрэглэгчийн хэлээр (mn/en) гарна

## 4. "Өөрийгөө удирдах" (Self-Managing) давхарга

Claude agent loop + хязгаарлагдсан tool registry:

1. **Tool registry** (Go): tool бүр нь interface — `Name()`, `Schema()`, `RequiredRole()`, `Execute(ctx)`. RLS context-оор дамжина, **AI ч гэсэн хэрэглэгчийн эрхээс хэтэрч чадахгүй**
2. **Эхний tool-ууд**: хэрэглэгчийн удирдлага (унших/идэвхгүй болгох — admin), metric унших, log хайх, имэйл загвар илгээх, тохиргоо унших (нууцгүй)
3. **Scheduled agent runs**: өдөр бүр health check → асуудал илэрвэл Claude дүн шинжилгээ хийж incident summary (MN/EN) үүсгэж админд имэйлдэнэ
4. **Human-in-the-loop**: уншихаас бусад бүх үйлдэл (устгах, идэвхгүй болгох, илгээх) админ баталгаажуулалт шаардана — approval queue UI
5. **Audit заавал**: AI-ийн tool call бүр `pkg/audit`-аар бичигдэнэ (одоо байгаа audit package-ээ яг энд бүрэн ашиглана)
6. **Prompt injection хамгаалалт**: хэрэглэгчийн оруулсан текстийг system prompt-оос тусгаарлах, tool үр дүнг untrusted гэж үзэх, output-д secret scan хийх

## 5. i18n стратеги (урьдчилсан нөхцөл!)

AI давхаргаас ӨМНӨ хийх ёстой:

1. **Frontend**: `next-intl` + `src/i18n/mn.json`, `en.json`. Бүх hardcode текстийг dictionary руу гаргана
2. **Backend**: `go-i18n`, `Accept-Language` header → алдааны мессеж mn/en
3. **AI-assisted translation pipeline**: Claude Haiku-аар `mn.json` → `en.json` орчуулгыг CI дээр автоматжуулж, хүн review хийнэ (энэ нь өөрөө платформын эхний AI feature болно)
4. AI хариултын хэл: хэрэглэгчийн `gerege.lang` тохиргоог system prompt-д дамжуулна

## 6. Өгөгдөл, аюулгүй байдал, кост

- **Шинэ хүснэгтүүд** (бүгд RLS-тэй): `ai_conversations`, `ai_messages`, `ai_tool_calls` (audit), `ai_usage` (token/cost metering), `kb_documents` + `kb_chunks` (pgvector)
- **Хэрэглэгч тус бүрийн rate limit + token budget**: Redis дээр өдрийн quota; хэтэрвэл 429 + ойлгомжтой мессеж. (Одоогийн IP-rate limiter-ийг user-based болгож өргөтгөнө — review-ийн №4 сул талыг давхар шийднэ)
- **Нууц түлхүүрүүд**: `ANTHROPIC_API_KEY`, `GEMINI_API_KEY` зөвхөн backend env-д; `.env.example`-д нэмэх; production-д secret manager
- **Streaming**: текст хариулт SSE-ээр, дуу WebSocket-оор
- **PII**: AI рүү явуулахын өмнө имэйл/утас redact хийх сонголт; log-д prompt бүрэн хадгалахгүй (hash + metadata)

## 7. Хэрэгжүүлэх дараалал (roadmap)

| Үе шат | Агуулга | Хугацаа (ойролцоо) |
|---|---|---|
| **0. Суурь засвар** | i18n (next-intl + go-i18n), error boundary, request-ID дамжуулалт, user-based rate limit, audit бүрэн холболт | 1–2 долоо хоног |
| **1. AI Gateway + Chat** | `internal/business/ai`, Claude SSE chat, conversation хадгалалт, usage metering, чат UI (mn/en) | 2–3 долоо хоног |
| **2. Self-Explaining** | pgvector RAG, docs/swagger индекс, платформын туслах, live metric tools | 2 долоо хоног |
| **3. Voice (Gemini)** | WebSocket proxy, STT/TTS, live mn⇄en voice translation UI | 2–3 долоо хоног |
| **4. Self-Managing** | Tool registry, scheduled agent, approval queue, incident summary | 3–4 долоо хоног |
| **5. Хатуужилт** | Load test, prompt injection red-team, cost dashboard, frontend тест | үргэлжилнэ |

## 8. Эхний алхам — ХЭРЭГЖСЭН ✅ (2026-06-04)

1. ✅ **i18n хоёр талдаа** — гадаад сан нэмэлгүй, төслийн minimal-dependency
   зарчмаар: backend `internal/i18n` (Accept-Language → mn/en, бүх API
   мессеж орчуулагдана), frontend `src/lib/i18n.ts` dictionary +
   `useT()` hook + `gerege.lang` cookie (server component-ууд хэл мэддэг
   боллоо). AppShell, нүүр, landing, login бүрэн mn/en.
2. ✅ **Backend AI модуль** — `internal/business/usecases/ai`,
   `pkg/aiclient` (Anthropic streaming клиент, SDK-гүй),
   `POST /api/v1/ai/chat` (SSE), `GET /api/v1/ai/conversations`,
   `GET /api/v1/ai/conversations/:id/messages`. Өдрийн хязгаар (Redis),
   токен метеринг (ai_usage), self-explaining system prompt.
3. ✅ **Migration 7** — ai_conversations / ai_messages / ai_usage,
   бүгд RLS + FORCE RLS бодлоготой (migration 6-ийн загвараар).
4. ✅ **Env** — `ANTHROPIC_API_KEY`, `ANTHROPIC_MODEL`, `AI_MAX_TOKENS`,
   `AI_DAILY_REQUEST_LIMIT`, `AI_REQUEST_TIMEOUT_SECS`, `GEMINI_API_KEY`
   (нөөц) — backend.env.example + internal/config/.env.example.
5. ✅ **Чат UI** — `/chat` хуудас (хамгаалагдсан route), SSE streaming
   BFF proxy (`/api/ai/chat`, токен cookie-д хэвээр), mn/en.

Шалгалт: frontend `tsc --noEmit` + `next lint` цэвэр. Backend-ийг
`make serve`-ээс өмнө `go build ./...` болон migration ажиллуулж шалгана
(`go run cmd/migration/main.go` загвараар). AI идэвхжүүлэхийн тулд
backend env-д ANTHROPIC_API_KEY тавихад л хангалттай.
