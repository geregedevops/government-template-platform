# Үе 1 — Федерацийн цөм: хэрэгжүүлэлтийн төлөвлөгөө

Эх төлөвлөгөө: `docs/WORLD-CLASS-ROADMAP.md` (Үе 1, 3–9 сар). Энэ баримт нь
Үе 1-ийг **бодит, дараалсан increment** болгож задална. Зорилго: root + 1 яам +
1 агентлаг **гурван тусдаа deploy** хоорондоо гарын үсэгтэй мессежээр ажиллах.

## Аль хэдийн бэлэн суурь (Үе 0-оос)
- **org tree + org_path_of + app.user_org** — "энэ платформ ямар байгууллагыг тээж байна" + scope.
- **ES256 + JWKS** (`pkg/jwt`, `/.well-known/jwks.json`) — асимметрик гарын үсэг + нийтийн түлхүүр түгээх загвар → М2М мессежийн гарын үсэгт ШУУД дахин ашиглана.
- **BPM engine** (`advance()` switch) + **instance snapshot/versioning** + `bpm_events` audit — delegatedTask-ийн суурь.
- **RLS withRLS** загвар, **outbox-д тохирох** repo+migration хэв маяг.

## Зарчим
1. **Гүйцэтгэлийг төвлөрүүлэхгүй** — node бүр дотроо orchestration; node хооронд гарын үсэгтэй мессеж + статус event (choreography).
2. **At-least-once + dedup = effectively-once** — outbox (sender) + processed-messages (receiver idempotency by `jti`).
3. **Trust = гарын үсэг + registry** — мессеж бүр илгээгчийн EC түлхүүрээр detached JWS; хүлээн авагч registry-ийн `jwks_url`-аас түлхүүрийг (kid-ээр) татаж шалгана. (Дараа: eID Mongolia CA + OCSP/CRL.)
4. **Гадны системийг mock-first** — eID, ХУР, ДАН-г интерфейсийн цаана mock-оор хийж, бодит интеграцийг блоклохгүй.
5. **Локалаар батлах** — протоколыг **loopback** (node өөрийгөө peer болгон бүртгэж round-trip) + дараа нь **3 docker compose instance**-аар турших; гадны систем шаардахгүй.

---

## Increment-ууд (дараалал)

### P1.1 — Node identity (federation e-seal) + fed-JWKS  ✅ хийгдэнэ
- Тусдаа **FED EC түлхүүр** (`FED_EC_PRIVATE_KEY`, P-256) — JWT-ээс ХАМААРАЛГҮЙ (live HS256 хэвээр). `pkg/fedsign` (ES256 detached JWS sign/verify + JWK).
- `/.well-known/fed-jwks.json` (BFF rewrite) — node-ийн нийтийн түлхүүр.
- Node-ийн өөрийн таних мэдээлэл: `FED_NODE_ID`, `FED_NODE_ORG` (тээж буй байгууллага).

### P1.2 — Platform registry (peers)  ✅ хийгдэнэ
- `fed_peers` хүснэгт: `id, key, name, org_id, base_url, jwks_url, status('pending'|'active'|'suspended'), kid, created_at`.
- CRUD (`/admin/fed/peers`, шинэ эрх `fed.manage`) + admin UI; onboarding (бүртгүүлэх → root батлах).

### P1.3 — Signed envelope + inbound endpoint  ✅ хийгдэнэ
- Envelope: `{ id(jti), typ, iss(node), aud(peer), iat, body }` + detached JWS (FED key, kid header).
- `POST /api/v1/fed/inbound` — peer-ийн гарын үсгийг (registry jwks_url, kid) шалгаж, `jti`-аар dedup (`fed_inbox` processed table), typ-ээр dispatch.
- **Loopback ping** (`typ=ping`) — node өөрийгөө active peer болгон бүртгэж, sign→send→verify→ack round-trip-ийг нэг deploy дээр батлана.

### P1.4 — Outbox + worker (durability)  ✅ хийгдэнэ
- `fed_outbox` хүснэгт: `id, peer_id, envelope, status, attempts, next_attempt_at, last_error`.
- Background worker (River/Asynq биш — энгийн polling worker, App lifecycle-д Stop()): backoff retry, N оролдлогын дараа dead-letter.
- Илгээх бүх мессеж outbox-оор → offline peer-ийг тэсвэрлэнэ.

### P1.5 — delegatedTask node + instance correlation  (дараагийн том increment)
- BPM graph-д `delegatedTask` (zeebe:taskHeaders: `delegate.peer`, `delegate.processKey`, input mapping).
- `bpm_process_instances`: `parent_instance_id`, `origin_peer`, шинэ status `waiting`.
- Engine `advance()`-д шинэ **"waiting" terminal state**: delegatedTask-д хүрэхэд `delegation.request` (outbox) илгээж instance-ийг `waiting` болгоно.
- Inbound `delegation.request` → дэд instance эхлүүлнэ (`origin_peer`, `parent_instance_id`); дуусахад `delegation.callback` (output vars) илгээнэ.
- Inbound `delegation.callback` → `waiting` parent-ийн хувьсагчдыг merge, delegatedTask-аас цааш advance (resume). Saga/compensation: callback `status=failed` → parent fail/compensate.

### P1.6 — Status events (telemetry) + dual monitoring
- Алхам бүр `status.event`-ийг гарал үүслийн node руу webhook-оор (outbox). Roll-up: node бүр бүрэн лог өөртөө, дээш зөвхөн metadata (X-Road op-monitoring загвар).

### P1.7 — eID Mongolia OIDC adapter (mock-first)
- `pkg/eid` интерфейс: OIDC нэвтрэлт (РД → push → PIN1), PIN2 remote signing. Эхэлж **mock provider** (локал тест), дараа e-id.mn бодит. ДАН OIDC хоёрдогч.

### P1.8 — ХУР client adapter (mock-first)
- serviceTask-ийн стандарт хэлбэр (developer.xyp.gov.mn WS каталог). Mock HTTP → бодит ХУР гэрээ хийгдсэний дараа.

### P1.9 — Пилот
- 1 бодит үйлчилгээ (ТӨҮГ-ийн зөвшөөрөл) root → яам → гүйцэтгэгч гинжээр 3 deploy дээр.

---

## Явц (live баталгаажсан, migration 24–25)
- ✅ **P1.1 Node identity** — FED EC түлхүүр (`pkg/fedsign`) + `/.well-known/fed-jwks.json`.
- ✅ **P1.2 Platform registry** — `fed_peers` + CRUD + `fed.manage` (migration 24).
- ✅ **P1.3 Signed envelope + inbound** — ES256 detached JWS, `/api/v1/fed/inbound`
  (нэвтрэлтгүй, peer JWKS-ээр kid-аар шалгана) + jti dedup (`fed_inbox`). **Loopback
  ping round-trip баталгаажсан.**
- ✅ **P1.4 Outbox + worker** — `fed_outbox` + backoff retry + dead-letter + 30с worker
  + `/fed/flush`.
- ✅ **P1.5 delegatedTask** — `delegate.peer`/`delegate.process` header (migration 25:
  parent_instance_id, origin_peer, 'waiting'); engine waiting-state + дэд instance +
  callback resume; bpm↔fed мөчлөгийг FedSender/DelegationHandler интерфейсээр таслав.
  **Loopback delegation round-trip баталгаажсан** (parent waiting → peer sub-instance →
  callback → parent completed). Бүх engine өөрчлөлт additive (зөвхөн delegate.peer node).

## Дараагийн session/гадны интеграц
- ⏭ **P1.6 Status events** — алхам бүр гарал үүсэл рүү webhook + dual monitoring.
- ⏭ **P1.7 eID Mongolia OIDC** (mock-first) — `pkg/eid` интерфейс + mock → e-id.mn бодит.
- ⏭ **P1.8 ХУР adapter** (mock-first) — serviceTask стандарт + mock → бодит ХУР.
- ⏭ **P1.9 Пилот** — 1 бодит үйлчилгээ 3 deploy дээр.

> P1.7/P1.8 гадны системийн эрх/гэрээ шаардана — интерфейс+mock-ийг урьдчилан хийнэ.
> Live JWT HS256 хэвээр (FED түлхүүр тусдаа).

## Эрсдэл/тэмдэглэл
- **FED түлхүүр live-д тохируулах** нь JWT-д нөлөөлөхгүй (тусдаа env) — хэрэглэгч дахин нэвтрэхгүй.
- **Loopback** нь протоколыг батлах хамгийн хямд арга — 3 deploy зөвхөн пилотод.
- **CA/OCSP (eID)** ирэхэд trust registry "түлхүүр өөрөө" → "CA-baked" болж хялбарчилна; одоогийн self-managed EC key нь bootstrap.
