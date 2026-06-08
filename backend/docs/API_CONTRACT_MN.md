# API Гэрээ (Contract)

> 🌐 [English](API_CONTRACT.md) · **Монгол**

**Gerege Template AI v1.0**-ийн REST API лавлагаа. Шууд, автоматаар үүсэх
бүрэн тодорхойлолтыг `GET /swagger/` дээр үзнэ (эх: `docs/swagger.json`).
Англи хувилбар: [API_CONTRACT.md](./API_CONTRACT.md).

> **Эх сурвалж.** Нээлттэй эх
> [snykk/go-rest-boilerplate](https://github.com/snykk/go-rest-boilerplate)
> (MIT, Najib Fikri)-аас гаралтай; HTTP давхаргыг **Gin → Fiber v3**,
> өгөгдлийн давхаргыг **sqlx → GORM** болгосон.

## Дүрэм

- **Үндсэн URL:** `http://localhost:8080/api/v1`
- **Content-Type:** `application/json`
- **Танилт:** хамгаалагдсан endpoint-д `Authorization: Bearer <access_token>` шаардана
- **Rate limit:** `/auth/*` нь IP тус бүр ~5 хүсэлт/минут (хэтэрвэл `429`)

### Хариуны бүтэц (envelope)

```json
{ "status": true, "message": "...", "data": { }, "request_id": "..." }
```
- `status` — амжилтад `true`, алдаанд `false`
- `data` — амжилтад л байна
- `request_id` — корреляцийн ID (`X-Request-ID` header-т мөн)

### Статус кодууд

| Код | Утга |
|-----|------|
| 200 / 201 | Амжилттай / Үүсгэсэн |
| 400 | Буруу body |
| 401 | Токен/нэвтрэлт буруу |
| 403 | Хориглосон (lockout) |
| 404 | Олдсонгүй |
| 409 | Давхцал (username/email) |
| 422 | Validation алдаа (`data.errors` дотор талбараар) |
| 429 | Хэт олон хүсэлт |
| 500 | Дотоод алдаа |

`strongpassword` дүрэм: том/жижиг үсэг, тоо, тусгай тэмдэгт + доод тал нь 12 тэмдэгт.

---

## Танилт (Authentication)

| Method | Path | Body | Амжилт (200/201) |
|--------|------|------|------------------|
| POST | `/auth/register` | `username`(3–25), `email`(≤50), `password`(12–72, strong) | `201` "registration user success" + user |
| POST | `/auth/login` | `email`, `password` | `200` "login success" + user + `token` + `refresh_token` |
| POST | `/auth/send-otp` | `email` | `200` "otp code has been send to …" |
| POST | `/auth/verify-otp` | `email`, `code`(numeric) | `200` "otp verification success" |
| POST | `/auth/refresh` | `refresh_token` | `200` "token refreshed" + шинэ token pair |
| POST | `/auth/logout` | `refresh_token` | `200` "logout success" |
| POST | `/auth/password/forgot` | `email` | `200` "if the email is registered…" |
| POST | `/auth/password/reset` | `token`, `new_password`(strong) | `200` "password reset" |
| PUT 🔒 | `/auth/password/change` | `current_password`, `new_password`(strong) | `200` "password changed" |

### Жишээ: нэвтрэх

**Хүсэлт** `POST /api/v1/auth/login`
```json
{ "email": "john@example.com", "password": "Str0ng!Passw0rd" }
```
**Хариу `200`**
```json
{ "status": true, "message": "login success", "data": {
  "id": "…", "username": "johndoe", "email": "john@example.com", "role_id": 2,
  "token": "<access_jwt>", "refresh_token": "<refresh_jwt>",
  "created_at": "…", "updated_at": null }, "request_id": "…" }
```
Алдаа: `401` нэвтрэлт буруу, `403` дараалсан амжилтгүйн дараа lockout.

---

## Хэрэглэгч (Users)

| Method | Path | Тайлбар |
|--------|------|---------|
| GET 🔒 | `/users/me` | Нэвтэрсэн хэрэглэгчийн профайл — `200` "user data fetched successfully" |

---

## AI туслах (Anthropic Claude) 🔒

`ANTHROPIC_API_KEY` хоосон бол бүгд `503`. Хэрэглэгчийн өдрийн лимиттэй.

| Method | Path | Тайлбар |
|--------|------|---------|
| POST 🔒 | `/ai/chat` | Streaming чат (**SSE**, JSON envelope биш). Body: `{ conversation_id?, message }`. Event: `delta` / `done` / `error`. |
| GET 🔒 | `/ai/conversations` | Ярианы жагсаалт (`offset`,`limit`≤50) |
| GET 🔒 | `/ai/conversations/{id}/messages` | Нэг ярианы бүх мессеж (эзэмшигч биш → `404`) |

## Дуу хоолой (Google Gemini) 🔒

`GEMINI_API_KEY` хоосон бол бүгд `503`. Аудио base64-аар (1 MiB body cap дотор);
буцах аудио нь WAV (PCM 24kHz/16/mono).

| Method | Path | Тайлбар |
|--------|------|---------|
| POST 🔒 | `/voice/translate` | MN↔EN яриа орчуулга (STT→орчуулга→TTS). Body: `{ source_lang: mn\|en, mime_type, audio_base64 }` → `{ source_text, translated_text, audio_base64, … }` |
| GET 🔒 | `/voice/history` | Орчуулгын түүх (аудиогүй; `offset`,`limit`≤50) |
| POST 🔒 | `/voice/transcribe` | Дуу→бичвэр (чатын микрофон). Body: `{ lang, mime_type, audio_base64 }` → `{ text }` |
| POST 🔒 | `/voice/speak` | Бичвэр→дуу (чатын "Сонсох"). Body: `{ text }` → `{ audio_base64, audio_mime }` |

---

## Үйлдлийн endpoint-ууд (`/api/v1` угтваргүй)

`GET /health` (liveness) · `GET /ready` (Postgres + Redis шалгана) ·
`GET /metrics` (Prometheus) · `GET /swagger/*` (Swagger UI)

---

🔒 = `Authorization: Bearer <access_token>` шаардана. Энэ тодорхойлолтыг handler
annotation-аас `make swag`-аар дахин үүсгэнэ.

---

**Gerege Template AI v1.0** — **Gerege Systems Development Team** болон **Claude AI** хамтран бүтээв, 2026.
