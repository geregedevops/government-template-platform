# Gerege Template → Монгол Улсын Засгийн газрын суурь платформ
## Хийрархит (root → яам → агентлаг → ТӨҮГ) платформуудын жишиг template болгох — Gap Analysis

Огноо: 2026-06-08 · Суурь: gerege-template-ai-bpm (одоогийн код)

---

## 0. Дүгнэлт (TL;DR)

Санаа **бүтцийн хувьд зөв** — root дээр иргэний хүсэлт → доош task болж задрах → доороо гүйцэтгэл → дээшээ батлагдаад нэгтгэгдэх → root бүгдийг хянах нь дэлхийн батлагдсан загвар (X-Road, India Stack, Camunda/Temporal-ийн distributed хувилбар).

Гэхдээ одоогийн код бол **нэг платформын (single-node) AI-BPM суурь** — нэг байгууллага дотор process үүсгэж, ажиллуулж, хянадаг. Хийрархит, федератив, олон-платформын давхарга нь **бараг бүхэлдээ дутуу**. Энэ нь муу гэсэн үг биш: суурь архитектур (Clean Architecture, RLS, BPM engine, AI, voice, RBAC) маш чанартай тул дээр нь хийрархит давхаргыг **нэмж** болно. Доор юу дутууг 10 capability domain-аар эрэмбэлэв.

Зэрэглэл: **P0** = энэгүйгээр хийрархит платформ боломжгүй · **P1** = жишиг template-д заавал · **P2** = боловсронгуй болгох.

---

## 1. Multi-tenancy / Байгууллагын мод (Organization hierarchy) — **P0, бүхэлдээ дутуу**

**Одоогийн байдал:** Код дотор `organization`, `tenant`, `agency`, `ministry`, `org_id` гэсэн ойлголт **огт байхгүй** (grep — 0 үр дүн). Бүх entity (`bpm_process_definitions`, `instances`, `tasks`, `ai_*`, `voice_*`) зөвхөн `user_id`-аар эзэмшигддэг. RLS бол **per-user**, per-organization биш.

**Жишиг платформд хэрэгтэй:**
- `organizations` хүснэгт — өөрийгөө лавладаг мод (`parent_id`), төрөл (`root` / `ministry` / `agency` / `soe`), түвшин, ОРОНГ/improvedcode.
- Бүх resource-д `org_id` (эзэмшигч байгууллага) нэмэх.
- RLS-ийг **org-scoped hierarchy**-болгох: `current_setting('app.org_path')`-аар "миний болон доорх бүх байгууллага"-г шүүх (Postgres `ltree` эсвэл materialized path тохиромжтой). Жишээ predicate: `org_path <@ current_setting('app.org_scope')::ltree`.
- Хэрэглэгч ↔ байгууллага холбоо (нэг хүн хэд хэдэн байгууллагад үүрэгтэй байж болно).

**Яагаад P0:** "Яам өөрийн доорхио хардаг, root бүгдийг хардаг" гэдэг бол яг л hierarchical org-scoped RLS. Үүнгүйгээр хийрархи гэж байхгүй.

---

## 2. Federated identity / Inter-platform trust (платформ хоорондын итгэл) — **P0, дутуу**

**Одоогийн байдал:** Auth нь нэг платформын дотоод — JWT (HS256, symmetric secret), refresh token Redis-д. **OIDC, OAuth2 client-credentials, SAML, mTLS, SSO холбоо алга** (grep — 0). Платформ хоорондоо хүсэлт дамжуулахдаа итгэлцэх механизм байхгүй.

**Жишиг платформд хэрэгтэй:**
- Платформ-хоорондын **machine-to-machine auth**: OAuth2 client-credentials эсвэл mTLS. Root яам руу task илгээхдээ өөрийгөө баталгаажуулах ёстой.
- **Asymmetric JWT (RS256/ES256)** болон JWKS endpoint — symmetric secret-ийг олон платформд хуваалцах нь аюултай. Одоо `pkg/jwt` нь HS256 hardcode (review-ийн H-аас).
- Иргэний нэвтрэлт нь үндэсний eID/ДАН (Монголын Digital ID) integration цэг — энэ нь template-д adapter interface байх ёстой.
- **Trust registry**: ямар платформ найдвартай, ямар public key-тэйг бүртгэх (X-Road-ийн trust service-ийн эквивалент).

**Яагаад P0:** Distributed гүйцэтгэл = платформууд бие биедээ итгэх ёстой. Дотоод symmetric JWT-ээр энэ боломжгүй.

---

## 3. Inter-platform routing & messaging (хүсэлт дамжуулах суваг) — **P0, дутуу**

**Одоогийн байдал:** BPM-ийн `serviceTask` нь `pkg/bpmconnector` — **зүгээр outbound HTTP** (GET/POST, header, body). SSRF хамгаалалт сайн, гэхдээ:
- **Async messaging / event bus алга** (NATS/Kafka/RabbitMQ — grep 0). Бүх дуудлага synchronous, request-time.
- **Webhook / callback receiver алга** — доод платформ гүйцэтгээд "буцааж дээшээ медэлдэг" суваг байхгүй.
- **Outbox / retry / idempotency алга** (grep 0) — дуудлага амжилтгүй болбол алдагдана.
- Connector нь request signing хийдэггүй (HMAC/mTLS алга) — хүлээн авагч илгээгчийг батлах боломжгүй.

**Жишиг платформд хэрэгтэй:**
- **Asynchronous task dispatch**: root → яам руу task илгээх нь message (durable queue), HTTP биш. Доод платформ offline байж болзошгүй.
- **Callback/result ingestion endpoint**: доод платформ гүйцэтгэлийн үр дүнг **signed callback**-аар буцаах. Correlation ID-аар анхны хүсэлттэй холбох.
- **Transactional outbox pattern** — process state өөрчлөлт + гадагш message нэг атом дотор (одоо `advance()` нь synchronous HTTP, partial failure-д эмзэг — review-ийн M3).
- **Idempotency key** бүх inbound task дээр — давхар дамжуулалтаас хамгаалах (review-д RunClient давхар-start H2 яг үүнтэй холбоотой шинж).
- Дамжуулалтын **signing**: HMAC эсвэл detached JWS.

**Яагаад P0:** "Доош хувиарлагдаж, гүйцэтгэл явж, дээшээ медэлдэг" гэсэн чиний гол урсгал нь яг энэ давхарга. Одоо зөвхөн synchronous fire-and-forget HTTP байна.

---

## 4. Distributed process / Cross-platform orchestration (хийрархит workflow) — **P0, хэсэгчилсэн**

**Одоогийн байдал:** BPM engine бий, гэхдээ **нэг платформ дотор хаалттай**. `bpm_process_instances.current_node_id` нь token байрлал хадгалдаг (сайн суурь), гэхдээ:
- Process graph нь нэг JSONB document, нэг engine дээр гүйцэтгэгдэнэ. **Өөр платформ дээрх sub-process-ыг дуудах node type алга.**
- `serviceTask` нь HTTP дуудна, гэхдээ "энэ алхмыг доод яамны платформ гүйцэтгэнэ, дуустал хүлээнэ, үр дүнг буцааж авна" гэсэн **inter-platform task node алга.**
- Saga / compensation (доод алхам бүтэлгүйтвэл дээд алхмуудыг буцаах) алга.

**Жишиг платформд хэрэгтэй:**
- Шинэ node type: **`delegatedTask`** — өөр платформын process-ыг дуудаж, callback хүлээж, үр дүнг variable-д merge хийдэг.
- **Correlation / process linking**: root instance ↔ доод платформын instance хооронд эцэг-хүү холбоо (`parent_instance_id`, `origin_platform`).
- **Distributed saga / compensation** — олон яам дамжсан хүсэлт алхамдаа бүтэлгүйтвэл буцаах logic.
- **Long-running / durable execution** — хүсэлт өдөр/долоо хоног үргэлжилж болно; одоо synchronous `advance()` нэг HTTP request дотор бүгдийг гүйцэтгэхийг оролддог (масштаблахгүй).

**Яагаад P0:** Энэ бол чиний санааны цөм — хийрархит делегаци. Engine-ийн суурь бий, гэхдээ distributed болгох node type, correlation, durability дутуу.

---

## 5. Hierarchical observability / Roll-up monitoring (хяналтын платформ) — **P1, дутуу**

**Одоогийн байдал:** `pkg/observability/metrics.go` (Prometheus), per-instance metric бий. `bpm_events` audit хүснэгт бий (сайн). Гэхдээ:
- **Cross-platform roll-up алга** — "яам бүрийн доорх бүх платформын хүсэлтийн төлөв"-ийг нэгтгэх view байхгүй.
- Root дээр "бүх системийн бүтсэн/бүтээгүй хүсэлт" харах **federated dashboard алга.**
- Доод платформоос төлөв татах/түлхэх стандарт **telemetry contract алга.**

**Жишиг платформд хэрэгтэй:**
- **Hierarchical status aggregation**: хүсэлт бүрийн төлөв (pending/in-progress/completed/failed/SLA-breached) org-modоор нэгтгэгдэж, дээд түвшин доод бүгдийг roll-up-аар хардаг.
- **SLA / deadline tracking** — төрийн үйлчилгээ хуулийн хугацаатай. Хугацаа хэтэрсэн хүсэлтийг улаанаар тэмдэглэх.
- **Federated metrics/trace contract** — доод платформ бүр стандарт форматаар (OpenTelemetry) төлөв тайлагнах.
- **Drill-down** — root-оос эцсийн нэгж хүртэл нэг хүсэлтийг мөрдөх (distributed trace, correlation ID-аар).
- Cowork artifact / live dashboard энд тохиромжтой.

**Яагаад P1:** Чиний "хяналтын платформ нь эцсийн нэгж хүртэл" гэдэг шаардлага. Суурь (events, metrics) бий, нэгтгэх давхарга дутуу.

---

## 6. Platform registry / Service catalog (платформ бүртгэл) — **P1, дутуу**

**Одоогийн байдал:** Бүртгэлийн ойлголт алга. "Одоо ажиллаж буй болон ирээдүйд бий болох бүх платформоо бүртгэж хянана" гэсэн чиний шаардлагыг биелүүлэх **registry байхгүй.**

**Жишиг платформд хэрэгтэй:**
- **Platform registry**: ямар платформ байгаа, эзэмшигч байгууллага, endpoint, public key, health, ямар үйлчилгээ (capability) санал болгодог.
- **Service catalog**: иргэнд санал болгох төрийн үйлчилгээний жагсаалт, аль платформ гүйцэтгэдэг, ямар маягт/process-той.
- **Onboarding flow**: шинэ яам/агентлаг өөрийн платформоо бүртгүүлэх (self-registration + root approval).
- **Health / heartbeat monitoring** доод платформуудын.

**Яагаад P1:** "Бүх платформоо бүртгэж хянана" = service registry + health monitoring. Template-ийн салшгүй хэсэг.

---

## 7. RBAC → ABAC / Hierarchical authorization — **P1, өргөтгөх**

**Одоогийн байдал:** RBAC бий (`rbac` feature, permission key, role) — сайн суурь. Гэхдээ role-based, **org-hierarchy-aware биш**. "Яамны admin зөвхөн өөрийн доорхио удирдана" гэдгийг одоогийн RBAC илэрхийлж чадахгүй (permission нь global).

**Жишиг платформд хэрэгтэй:**
- **Scoped roles**: "X яамны admin", "Y агентлагийн operator" — эрх нь org-subtree-д хязгаарлагдана.
- **ABAC элемент**: нөхцөлт эрх (зөвхөн өөрийн org_path доторх resource).
- **Delegation**: дээд түвшин эрхээ доош түр шилжүүлэх.
- Энэ нь #1 (org hierarchy)-тэй шууд хосолно.

**Яагаад P1:** Хийрархит эрх байхгүй бол хийрархит хяналт хэрэгжихгүй.

---

## 8. Audit, compliance, хууль эрх зүй — **P1, хэсэгчилсэн**

**Одоогийн байдал:** `bpm_events` audit бий. RLS, secret hygiene сайн. Гэхдээ төрийн платформд:
- **Tamper-evident audit log алга** (hash chain / append-only WORM) — төрийн шийдвэр маргаангүй байх ёстой.
- **Persons' data protection** (Хувийн мэдээллийн хамгаалалтын тухай хууль) — consent, retention, эрхээ эдлэх (хандах/устгах) flow алга.
- **Цахим гарын үсэг** (eSignature) — албан ёсны шийдвэрт хэрэгтэй; одоо алга.
- **Бүртгэлийн хадгалалтын хугацаа** (records retention) policy алга.

**Яагаад P1:** Төрийн платформ хууль зүйн нотлох чадвартай байх ёстой.

---

## 9. Resilience, масштаб, гүйцэтгэл — **P1/P2, дутуу**

**Одоогийн байдал (review-ээс):**
- Rate limiter, BPM advance() нь **per-instance / synchronous** — олон replica дээр масштаблахгүй (review M-ууд).
- Worker / background job system алга — урт хугацааны process-ыг async гүйцэтгэх дэд бүтэц байхгүй.
- Idempotency, retry, dead-letter queue алга (#3-тай давхцана).

**Жишиг платформд хэрэгтэй:**
- **Durable job/worker queue** (Temporal / Asynq / River гэх).
- **Horizontal scale**: stateless API, Redis-backed rate limit, distributed lock.
- **Circuit breaker / bulkhead** доод платформ унавал root-ыг хамгаалах.
- **Backpressure** — нэг яам удаашрахад бусдыг блокдохгүй.

---

## 10. Template-ийн боловсронгуй байдал (DX / reusability) — **P2**

Жишиг template болгохын тулд:
- **Scaffolding / generator**: шинэ яам платформ үүсгэх CLI (`scripts/rename-module.sh` бий — сайн эхлэл).
- **Adapter pattern**: ДАН/eID, eSignature, төлбөрийн систем, мессеж bus-ийг **plug-in interface** болгох (хэрэгжилт солигдоно).
- **Reference deployment**: root + 1 яам + 1 агентлаг-ийг харуулсан жишээ compose/k8s.
- **Conformance test suite**: шинэ платформ "template-д нийцэж байна уу"-г шалгах гэрээ-тест.
- **Versioned API contract** (OpenAPI/protobuf) платформ хооронд.
- Баримтжуулалт: integration guide, trust onboarding, observability contract.

---

## Тэргүүлэх дараалал (roadmap)

**Үе 1 — Хийрархийн суурь (P0):**
1. Organization hierarchy + org-scoped RLS (#1)
2. Federated identity: asymmetric JWT + JWKS + M2M auth (#2)
3. Async messaging + signed callback + outbox/idempotency (#3)
4. `delegatedTask` node + instance correlation + durability (#4)

**Үе 2 — Хяналт ба бүртгэл (P1):**
5. Platform registry + service catalog + health (#6)
6. Hierarchical roll-up dashboard + SLA tracking (#5)
7. Scoped RBAC/ABAC (#7)
8. Tamper-evident audit + eSignature + ХМ хамгаалалт (#8)

**Үе 3 — Бат бөх ба template (P1/P2):**
9. Durable worker queue + horizontal scale (#9)
10. Adapter plug-ins + scaffolding + conformance tests + reference deploy (#10)

---

## Нэг өгүүлбэрээр

Одоогийн код бол **маш чанартай нэг-платформын AI-BPM суурь** — хийрархит федератив төрийн платформын **доод нэг зангилаа (node)** болохуйц. Жишиг template болгохын тулд гол дутуу нь дөрвөн P0 давхарга: **(1) байгууллагын мод + org-scoped RLS, (2) платформ хоорондын итгэл/identity, (3) асинхрон дамжуулалт + callback, (4) платформ хоорондын delegated workflow.** Энэ дөрвийг нэмбэл одоогийн суурь нь жинхэнэ "root → яам → агентлаг → ТӨҮГ" хийрархийг тээх чадвартай болно.
