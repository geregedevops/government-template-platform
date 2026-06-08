# Дэлхийн түвшинд хүргэх хөгжүүлэлтийн төлөвлөгөө
## Gerege Template → Монгол Улсын төрийн үйлчилгээний хийрархит платформын жишиг

Огноо: 2026-06-08 · Судалгааны эх сурвалж: 2024–2026 оны албан ёсны баримтууд (доор иш татав)

---

## 1. Хураангуй

Дэлхийн тэргүүлэгчдийг (Эстони X-Road, Украин Diia, Энэтхэг India Stack, Сингапур SGTS, ЕХ OOTS/eIDAS 2.0, Их Британи GDS, GovStack) 2024–2026 оны байдлаар судлахад **нэг тогтсон загвар** давтагдаж байна:

1. **Нэг үндсэн хаалга** (eesti.ee, Diia, GOV.UK) — үйлчилгээг угсарч, төлвийг нэгтгэн харуулна.
2. **Федератив солилцооны давхарга** — байгууллага бүр өөрийн gateway node ажиллуулна (X-Road Security Server, eDelivery Access Point); мессеж бүр гарын үсэгтэй, лог-той, цаг-тамгатай. Төв нь зөвхөн гишүүдийн бүртгэл + итгэлийн язгуур + үйлчилгээний каталог хадгална.
3. **Workflow/BPMN engine нь байгууллага бүрийн ДОТОР** (Diia.Engine, Бельгийн FPS BOSA-гийн Camunda, GovStack Workflow BB) — байгууллага хоорондын алхмыг нэг том engine биш, **гарын үсэгтэй мессеж + статус event** солилцоогоор хийнэ.
4. **Хоёр давхар мониторинг** — бүрэн лог нь node бүр дээрээ үлдэж, төв нь зөвхөн operational metadata/SLA статистик хардаг (data sovereignty хадгалагдана).
5. **Төвийг сахисан core-software steward** (NIIS, NPCI, GovTech, Еврокомисс) хувилбартай release гаргаж, гишүүд **өөрсдийн хуваарьт цонхонд** өөрсдөө шинэчилнэ.

Энэ нь бидний өмнөх дүгнэлтийг (нэг versioned core + N бие даасан платформ + нэг протокол) дэлхийн практикаар **бүрэн баталж** байна. Хамгийн ойрын аналог нь **Украины Diia.Engine**: 20 яам/агентлаг хэрэглэдэг, 50+ бүртгэл, 100+ үйлчилгээг тээдэг low-code engine, доороо Trembita (X-Road суурьтай) bus-тай. **Gerege Template-ийн зорилт = "Монголын Diia.Engine" байх** гэж томъёолж болно.

Монголд аль хэдийн **ХУР** (солилцооны давхарга: 178 төрийн + 481 хувийн байгууллага, 4.6 тэрбум солилцоо) болон **ДАН** (2.1 сая хэрэглэгч, SSO + тоон гарын үсэг) ажиллаж байгаа тул template нь эдгээрийг **дахин бүтээх биш, залгах** ёстой — энэ нь дэлхийн жишгээс гарах хамгийн чухал локал дүгнэлт.

Үүн дээр нэмж, **өөрсдийн бүтээсэн eID Mongolia (e-id.mn) бэлэн болсон** нь стратегийн гол хөзөр: Эстонийн Smart-ID загварыг дагасан, лицензтэй CA бүхий итгэлцлийн үйлчилгээ (PIN1 нэвтрэлт, PIN2 хуулийн хүчинтэй гарын үсэг, цахим тамга, qualified timestamp, OCSP/CRL, HSM). Энэ нь федерацийн итгэлцлийн давхаргыг (P0) гаднаас хүлээхгүйгээр **in-house шийдэх** боломж — Gerege (engine) + eID Mongolia (trust) хослол нь Украины Diia.Engine + Diia.Signature-ийн шууд эквивалент болно.

---

## 2. Дэлхийн жишгийн харьцуулсан зураглал

| Хэмжүүр | Эстони X-Road | Украин Diia | Энэтхэг | Сингапур | ЕХ OOTS | UK GDS | GovStack |
|---|---|---|---|---|---|---|---|
| Загвар | Федератив peer mesh | Төв портал + low-code engine + федератив bus | Төв utility + федератив гишүүд | Бүрэн төвлөрсөн WOG PaaS | Улсуудын федераци + нийтлэг routing | Төв портал + хуваалцсан компонент | Building Block спек |
| Routing | Point-to-point Security Server | Diia.Engine → Trembita → бүртгэлүүд | NPCI/UIDAI төв switch | APEX төв gateway | Evidence Broker + DSD → AS4 async | Шууд API | Information Mediator (X-Road + pub/sub) |
| Итгэл | PKI, e-seal, timestamp | BankID/QES + Trembita PKI | PKI, eSign, consent | Singpass | Qualified trust + signed AS4 | One Login (OIDC) | X-Road + Consent BB |
| Масштаб | 1 тэрбум query/жил, 450+ байгууллага | 22 сая хэрэглэгч, 14 тэрбум солилцоо | 20 тэрбум төлбөр/сар | 5 сая хэрэглэгч, 100 сая API/сар | 27 улс | 12 тэрбум мессеж | 6+ улсын пилот |
| Нээлттэй эх | MIT | EUPL-1.2 | MOSIP MPL-2.0 | Хаалттай | EUPL/Apache ref-impl | MIT | Нээлттэй спек |
| Steward | NIIS (3 улсын яамд) | MinDigital | NPCI/UIDAI | GovTech | Еврокомисс | GDS | ITU/GIZ/Эстони |

**UN EGDI 2024:** Дани (1), Эстони (2), Сингапур (3), БНСУ (4). **Монгол: 193-аас 46-д** — 2022 оноос 28 байр ахиж "Very High" ангилалд анх удаа орсон. Үндэслэл: e-Mongolia нэгтгэл + 2022 оны хуулийн багц.

---

## 3. Монголын одоогийн орчин — template хаана байрлах вэ

| Дэд бүтэц | Төлөв (2025–2026) | Template-ийн харьцаа |
|---|---|---|
| **e-Mongolia** (И-Монгол Академи) | 2+ сая хэрэглэгч, 87 байгууллагын 1,260+ үйлчилгээ, v5.0 (AI чиглэл) | Иргэний хаалга аль хэдийн бий — template нь **байгууллага талын гүйцэтгэлийн engine** байр эзэлнэ |
| **ХУР** (Үндэсний Дата Төв) | ESB архитектур, 178 төрийн + 481 хувийн байгууллага, 4.6 тэрбум солилцоо, developer.xyp.gov.mn | Солилцооны давхаргыг **дахин бүтээхгүй** — ХУР adapter бичнэ. Урт хугацаанд ХУР-ыг X-Road төст федератив загварт ойртуулах нөлөө үзүүлж болно |
| **ДАН** (sso.gov.mn) | 2.1 сая хэрэглэгч, ДАН 2.0 + тоон гарын үсэг модуль, 98 төрийн + 276 хувийн байгууллагын системүүд | Иргэн/албан хаагчийн нэвтрэлтийг **ДАН OIDC adapter**-аар; e-mail+OTP нь fallback |
| **eID Mongolia** (e-id.mn — өөрсдийн бүтээл, бэлэн) | Лицензтэй CA; Smart-ID загвар: PIN1 нэвтрэлт / PIN2 гарын үсэг, апп + desktop + USB токен; OAuth2/OIDC, REST API, webhook, mobile SDK; remote/batch signing, цахим тамга (e-seal), qualified timestamp, OCSP/CRL, HSM | **Итгэлцлийн давхаргын үндсэн шийдэл** — доорх 3.1-р хэсэгт дэлгэрэнгүй |
| **Тоон гарын үсэг (төрийн)** | 2022 оны Цахим гарын үсгийн хууль, esign.gov.mn, иргэний үнэмлэхний чип; 2025-д timestamp service зорилт | eID Mongolia-гийн зэрэгцээ хүлээн зөвшөөрөгдсөн альтернатив суваг |
| **Хуулийн орчин** | 2022.05.01: Нийтийн мэдээллийн ил тод байдал, Хүний хувийн мэдээлэл хамгаалах, Цахим гарын үсэг, Кибер аюулгүй байдлын хуулиуд | Consent, retention, audit шаардлагууд template-д кодлогдох ёстой |
| **Бодлого** | "Digital First": 90% үйлчилгээ онлайн 24/7 (2026–2030); Үндэсний AI/Big Data стратеги (2025) | Template-ийн AI давхарга (chat, voice, AI-BPM) бодлоготой шууд нийцнэ |
| **Хурдан** (khurdan.gov.mn) | 577 үйлчилгээний цэг, 482 KIOSK | Hurdan платформын нэршил эндээс; офлайн цэгүүд нь template-ийн front-end суваг болж болно |

**Дутагдал (ОУ-ын үнэлгээгээр):** Дэлхийн банк, АХБ хоёулаа "агентлагууд тус тусдаа силостой систем хөгжүүлсээр байгаа, whole-of-government interoperability давхарга дутуу" гэж дүгнэсэн. **Энэ цоорхой нь яг Gerege Template-ийн эзлэх байр.**

### 3.1 eID Mongolia — бэлэн итгэлцлийн давхарга (интеграцийн 5 цэг)

Эстонид X-Road-ийн итгэлцэл нь PKI/e-seal/timestamp дээр, иргэний гарын үсэг нь Smart-ID (SK ID Solutions) дээр тогтдог. eID Mongolia энэ хоёр үүргийг хоёуланг нь Монголд гүйцэтгэх чадвартай тул gap analysis-ийн P0 #2 (федератив итгэл) болон P1 #8 (audit/eSignature)-ийн дийлэнхийг **бэлэн системээр** хаана:

1. **Хүний нэвтрэлт** — иргэн/албан хаагч бүх tier-ийн платформд eID OIDC-ээр (РД оруулах → апп-д push → PIN1) нэвтэрнэ. Нэг identity бүх хийрархид — Singpass-ийн "нэг ID, 700 байгууллага" загвар.
2. **Шийдвэрийн гарын үсэг** — BPM-ийн батлах (approval) task бүр PIN2 remote signing-ээр баталгаажина: татгалзах боломжгүй (non-repudiation), шүүхийн ач холбогдолтой. "Яамны дарга баталсан" гэдэг нь криптографаар нотлогдоно.
3. **Байгууллагын e-seal — M2M мессежийн гарын үсэг** — платформ хоорондын delegation/callback мессежийг байгууллагын цахим тамгаар (CA-гаас олгосон org сертификат) гарын үсэглэнэ. Trust registry нь өөрөө түлхүүр тараахын оронд **eID Mongolia CA + OCSP/CRL дээр тулгуурлана** — X-Road-ийн e-seal загвар, гэхдээ өөрсдийн CA-тай.
4. **Qualified timestamp** — audit hash-chain (tamper-evident лог)-ийн гинж бүрийг eID-ийн цаг тэмдэглэгээгээр бататгана — Эстонийн "лог бүр нотлох баримт" зарчим.
5. **eKYC/онбординг** — шинэ платформын хэрэглэгч, ТӨҮГ-ийн ажилтны бүртгэлийг eKYC-ээр шууд баталгаажуулна.

Үр дагавар: протоколын spec-ийн trust хэсэг "өөрсдөө түлхүүр удирдах" биш "CA-д суурилсан" болж хялбарчлагдана; Үе 1-ийн ажил багасна.

### 3.2 Олон сувгийн орц — иргэний хүсэлт хаанаас ч ирж болно

Иргэний хүсэлт **зөвхөн e-Mongolia-аас биш**: хувийн хэвшлийн платформоос (жишээ: gerege.mn), Хурдан цэг/KIOSK-оос, платформуудын өөрсдийн portal-оос ч орж ирнэ. Дэлхийн жишиг үүнийг баталдаг:

- **India Stack / UPI загвар**: NPCI-ийн нэг switch дээр GPay, PhonePe гэх хувийн апп-ууд ажиллаж 20 тэрбум гүйлгээ/сар хүргэдэг — төр switch-ээ, хувийн хэвшил UX-ээ хариуцдаг.
- **ХУР-ын прецедент**: Монголд аль хэдийн 481 хувийн байгууллага (банк, операторууд) ХУР-д холбогдсон — хувийн суваг нээх нь эрх зүй, практикийн хувьд танил зам.
- **Singpass API загвар**: 700 байгууллагын 2,000+ үйлчилгээ нэг identity-гээр.

Архитектурын шаардлага (root платформын **Channel API**):

- Суваг бүр **бүртгэлтэй, гэрээт channel partner** (channel registry — platform registry-ийн нэг төрөл), org e-seal-ээр мессежээ гарын үсэглэнэ.
- Иргэний таниулалт суваг үл хамааран **заавал eID Mongolia-аар** (PIN1) — суваг нь UX, итгэл нь eID. Ингэснээр gerege.mn-ээс ирсэн хүсэлт e-Mongolia-аас ирсэнтэй адил эрх зүйн хүчинтэй.
- Хүсэлтийн төлөв (status event) суваг руугаа webhook-оор буцаж мэдэгдэнэ — иргэн аль сувгаас өгсөн, тэндээсээ хянана.
- Зөвшөөрлийн (consent) бүртгэл: иргэн ямар сувгаар ямар мэдээллээ дамжуулахыг зөвшөөрснөө eID-ээр баталгаажуулна (India Stack-ийн consent artefact загвар).

---

## 4. Бидний платформ ↔ дэлхийн түвшин: capability харьцуулалт

| Capability (дэлхийн жишиг) | Тэргүүлэгчид | Gerege одоо | Зай |
|---|---|---|---|
| Байгууллага доторх BPMN engine | Diia.Engine, Camunda | ✅ Бий (AI-generatable, React Flow modeler) | Бага — durability, versioning дутуу |
| AI давхарга (chat, voice, AI-assisted) | Diia.AI, Bürokratt, e-Mongolia 5.0 | ✅ Бий (chat, voice, AI-BPM gen) | **Бага — энэ бол давуу тал** |
| Аюулгүй байдлын суурь (RLS, JWT, CSP) | бүгд | ✅ Сайн (review-ээр батлагдсан) | Бага |
| Байгууллагын мод + org-scoped эрх | бүгд | ❌ Алга | **Том (P0)** |
| Федератив identity (org-хоорондын M2M) | X-Road PKI, eIDAS | ❌ Template-д алга, гэхдээ **eID Mongolia CA/e-seal бэлэн** | **Дунд болж буурав (P0, adapter л үлдсэн)** |
| Гарын үсэгтэй async мессеж + callback | eDelivery AS4, IM pub/sub | ❌ Алга (sync HTTP л бий) | **Том (P0)** |
| Байгууллага хоорондын process choreography | OOTS 4-corner, Trembita | ❌ Алга | **Том (P0)** |
| Платформ/үйлчилгээний registry | X-Road Central Server, Evidence Broker+DSD | ❌ Алга | Том (P1) |
| Хоёр давхар мониторинг (local лог + төв metadata) | X-Road op-monitoring, WOGAA | ❌ Хэсэгчилсэн (bpm_events, Prometheus per-instance) | Дунд (P1) |
| eID Mongolia adapter (OIDC, PIN2 signing, e-seal, timestamp) | Smart-ID/SK загвар | ❌ Adapter алга, **систем нь бэлэн** | Дунд (P0-P1) |
| ХУР adapter | (Монголын онцлог) | ❌ Алга | Том (P1) |
| Олон сувгийн Channel API (e-Mongolia, gerege.mn, KIOSK) | UPI/PSP, Singpass API | ❌ Алга | Дунд (P1) |
| Steward governance + versioned release | NIIS загвар | ❌ Алга (нэг репо) | Дунд (P1) |
| Нээлттэй эх + conformance test | MIT/EUPL + GovStack | ❌ Алга | Дунд (P2) |
| Proactive / life-event үйлчилгээ | Эстони (төрөлтөд автомат тэтгэмж) | ❌ Алга | Урт хугацаа (P2) |

---

## 5. Хөгжүүлэлтийн төлөвлөгөө (36 сар, 4 үе шат)

### Үе 0 — Суурь бэхжүүлэлт (0–3 сар)
*Зорилго: нэг node-ийн чанарыг production түвшинд аваачих.*

- Review-ийн үлдсэн асуудлуудыг хаах (H7 i18n, M1/M2/M5/M6/M8/M9/M11/M12, LOW-ууд).
- **Байгууллагын мод**: `organizations` (parent_id, ltree path, төрөл: root/ministry/agency/soe) + бүх resource-д `org_id` + org-scoped RLS (`org_path <@ app.org_scope`).
- **Asymmetric JWT (ES256) + JWKS endpoint** — федерацийн урьдчилсан нөхцөл.
- BPM engine-ийн durability: synchronous `advance()`-ыг durable job queue (River/Asynq) руу шилжүүлэх, process definition versioning (publish-new-version урсгал).
- Гарц: production-ready нэг node, org-aware RLS, хувилбарласан BPM.

### Үе 1 — Федерацийн цөм (3–9 сар)
*Зорилго: root + 1 яам + 1 агентлаг гурван тусдаа deploy хоорондоо ажиллана.*

- **Протокол spec v0 (тусдаа репо)**: delegation request, signed callback, status event, telemetry — OOTS-ийн 4-corner AS4 ба X-Road мессежийн загвараас санаа авч JSON/REST дээр detached JWS гарын үсэгтэй.
- **M2M auth**: байгууллагын **e-seal (eID Mongolia CA-гаас олгосон org сертификат)** + OCSP/CRL шалгалт; trust registry нь түлхүүр тараахын оронд CA-д тулгуурласан гишүүнчлэлийн бүртгэл болно (X-Road Central Server-ийн хялбаршуулсан, CA-backed эквивалент). OAuth2 client-credentials нь session давхарга.
- **eID Mongolia OIDC adapter**: иргэн/албан хаагчийн нэвтрэлт (РД → push → PIN1) бүх tier дээр; ДАН OIDC нь хоёрдогч/fallback суваг.
- **Async давхарга**: durable queue + transactional outbox + idempotency key + dead-letter; offline платформыг тэсвэрлэх retry.
- **`delegatedTask` node type**: BPM графт "энэ алхмыг X платформ гүйцэтгэнэ" — instance correlation (`parent_instance_id`, `origin_platform`), callback-аар variables merge, saga/compensation.
- **ХУР client adapter** (лавлагаа татах serviceTask-ийн стандарт хэлбэр, developer.xyp.gov.mn-ийн WS каталогт суурилсан).
- **PIN2 approval signing**: BPM-ийн батлах task-уудад eID Mongolia remote signing — шийдвэр бүр хуулийн хүчинтэй гарын үсэгтэй.
- **Пилот**: 1 бодит үйлчилгээг (жишээ: нэг ТӨҮГ-ийн зөвшөөрлийн процесс) root → яам → гүйцэтгэгч гинжээр бүрэн гүйлгэх.
- Гарц: ажиллаж буй 3-node федераци + протокол v0.

### Үе 2 — Хяналт, бүртгэл, тусгаарлал (9–18 сар)
*Зорилго: "бүх платформоо бүртгэж, эцсийн нэгж хүртэл хянана" гэдгийг бодит болгох.*

- **Platform registry + service catalog**: гишүүн платформ, endpoint, түлхүүр, health/heartbeat, санал болгож буй үйлчилгээнүүд; onboarding урсгал (бүртгүүлэх → root батлах). Evidence Broker + DSD загварын semantic routing.
- **Хоёр давхар мониторинг**: node бүр бүрэн лог өөртөө; дээш зөвхөн operational metadata (тоо, хугацаа, амжилт/бүтэлгүй, SLA) — X-Road op-monitoring протоколын загвараар. Root болон яам бүрд **roll-up dashboard** (хүсэлт бүрийг drill-down хийж эцсийн нэгж хүртэл мөрдөх — correlation ID-аар).
- **SLA engine**: үйлчилгээ бүрд хуулийн хугацаа, хэтрэлтийн escalation.
- **Scoped RBAC/ABAC**: "X яамны admin зөвхөн өөрийн subtree" — org_path-д суурилсан эрх.
- **Tamper-evident audit** (hash-chain) + **eID Mongolia qualified timestamp**-аар гинж бүрийг бататгах; esign.gov.mn нь хүлээн зөвшөөрөгдсөн альтернатив.
- **Channel API + channel registry**: e-Mongolia, хувийн платформ (gerege.mn г.м), Хурдан KIOSK-ийг бүртгэлтэй суваг болгон холбох — иргэний таниулалт суваг үл хамааран eID-ээр, төлөв суваг руу webhook-оор буцна (3.2-р хэсэг).
- **Core-ийг салгах**: versioned Go module + npm package + протокол spec репо; платформ бүр өөрийн репотой болж core-оо import хийнэ. Scaffolding CLI (`gerege new platform --tier=ministry`).
- **Conformance test suite v1** — шинэ платформ протоколд нийцэж буйг батлах (GovStack-ийн conformance загвараар).
- Гарц: 2–3 яамны бодит хэрэглээ, тусдаа репо/deploy бүхий core+protocol governance.

### Үе 3 — Үндэсний масштаб ба дэлхийн түвшин (18–36 сар)
*Зорилго: тэргүүлэгчдийн ялгарах чадварт хүрэх.*

- **Steward байгууллагажуулалт**: NIIS загвараар core-ийн release-ийг хариуцах нэгж (МДДХХЯ-ны харьяа эсвэл И-Монгол Академи түшиглэсэн); шинэчлэлтийн mandatory цонх, LTS бодлого, security patch түгээлт.
- **e-Mongolia integration**: иргэний хүсэлт e-Mongolia-аас орж ирэх API gateway; төлөв буцааж e-Mongolia дээр харагдах.
- **GovStack Building Block нийцэл**: Workflow BB + Information Mediator BB спект conformance — олон улсын танигдал, донор санхүүжилтийн боломж (Дэлхийн банкны Smart Government II, АХБ-ны TA-тай уялдуулах).
- **AI давхаргыг гүнзгийрүүлэх** (Diia.AI, e-Mongolia 5.0 чиг хандлага): AI агент хүсэлтийг эхнээс нь дуустал хөтлөх, process-ийг AI санал болгох — одоогийн AI chat/voice/AI-BPM суурь энд **дэлхийтэй зэрэгцэх давуу тал**.
- **Proactive үйлчилгээ**: life-event trigger (Эстонийн "төрөлт бүртгэгдмэгц тэтгэмж автоматаар" загвар) — ХУР-ын бүртгэлийн event-ээс process автоматаар эхлэх.
- **Нээлттэй эх болгох** (MIT эсвэл EUPL-1.2, Diia-гийн жишгээр) — экосистем, итгэл, олон улсын экспортын боломж.
- Гарц: үндэсний хэмжээний federated mesh; EGDI-д нөлөөлөхүйц цахим гүйцэтгэлийн давхарга.

---

## 6. Амжилтын хэмжүүр (KPI)

| Хэмжүүр | 12 сар | 24 сар | 36 сар | Жишиг |
|---|---|---|---|---|
| Холбогдсон платформ (тусдаа deploy) | 3 (пилот) | 8–12 | 25+ | X-Road 450+ байгууллага |
| Гинжээр гүйцэтгэсэн үйлчилгээ | 1–3 | 15+ | 50+ | Diia.Engine 100+ үйлчилгээ |
| Хүсэлтийн end-to-end мөрдөлт (drill-down) | пилотод | бүх платформд | бүх платформд | X-Road op-monitoring |
| SLA хэтрэлтийн ил тод байдал | — | root dashboard | олон нийтэд нээлттэй статистик | shilen.gov.mn зарчим |
| Core release cadence | quarterly | quarterly + LTS | quarterly + LTS | NIIS загвар |
| Протоколын backward compatibility | v0→v1 | v1 тогтвортой | v1.x | X-Road 7→8 backwards-compatible |

---

## 7. Гол эрсдэл ба хариу

- **ХУР/ДАН-ийн хамаарал**: гадны системийн API өөрчлөлт, гэрээний процесс удаан → adapter-ийг тусгаарласан interface-тэй бичиж, mock-оор хөгжүүлэлтийг блоклохгүй байх; МДДХХЯ/ҮДТ-тэй албан ёсны түншлэл эрт эхлүүлэх. (Итгэлцлийн давхарга нь eID Mongolia — in-house тул энэ эрсдэлээс гарсан.)
- **eID Mongolia-гийн эрх зүйн хүлээн зөвшөөрөлт**: төрийн системд хувийн CA-гийн гарын үсгийг бүх яам хүлээн зөвшөөрөх ёстой → Цахим гарын үсгийн хуулийн лицензийн хүрээнд МДДХХЯ-тай албан баталгаажуулалт, esign.gov.mn-тэй cross-recognition эрт шийдэх.
- **Хувийн сувгийн (gerege.mn г.м) эрсдэл**: суваг иргэний нэрийн өмнөөс хүсэлт үүсгэх эрсдэл → суваг хэзээ ч иргэнийг төлөөлөхгүй, таниулалт + зөвшөөрөл нь заавал иргэний eID PIN-ээр; суваг зөвхөн гэрээт, e-seal-тэй, audit-тай.
- **Захиргааны өөрчлөлт** (сайд, бодлого солигдох) → GovStack/Дэлхийн банкны хүрээнд байршуулж, нэг яамнаас үл хамаарах steward бүтэц рүү яаралтай шилжих.
- **Нэг engine-д бүх яамыг оруулах уруу таталт** → дэлхийн жишиг: байгууллага хоорондоо choreography, дотроо orchestration. Протоколыг л төвлөрүүлж, гүйцэтгэлийг хэзээ ч төвлөрүүлэхгүй.
- **Аюулгүй байдлын patch түгээлт N репод** → core-ийг library болгож (Үе 2), платформууд хувилбар ахиулах mandatory цонхтой байх (NIIS загвар).

---

## 8. Эх сурвалж (гол)

- X-Road: x-road.global/architecture; docs.x-road.global (Security/Operational Monitoring); NIIS X-Road 8 status (niis.org/blog, 2025.06); github.com/nordic-institute/X-Road (MIT)
- Украин: Diia.Engine — kmu.gov.ua, oecd-opsi.org/innovations/diia-engine; Trembita — cyber.ee case study (150+ registries, 14B exchanges); opensource.diia.gov.ua (EUPL-1.2); digitalstate.gov.ua (Diia.AI)
- GovStack: specs.govstack.global; bb-information-mediator (GitHub); govstack.global/global-showcase
- ЕХ: OOTS High-Level Architecture (March 2025, ec.europa.eu digital-building-blocks); eIDAS 2.0 / ARF 2.4.0; Camunda + FPS BOSA press release
- Энэтхэг: indiastack.org; mosip.io (Нигери $83M, 26 улс); NPCI/UIDAI статистик
- Сингапур: developer.tech.gov.sg (SGTS, APEX 100M calls/mo); MDDI Singpass factsheet
- UK: sign-in.service.gov.uk (One Login 122 services); notifications.service.gov.uk (12B messages); github.com/alphagov
- Монгол: mddic.gov.mn (Digital First, 2026–2030, ДАН 2.0, тоон гарын үсэг); datacenter.gov.mn/service/xyp (ХУР: 178+481 байгууллага, 4.6B); ema.gov.mn; khurdan.gov.mn (577 цэг); montsame.mn (EGDI 46); Дэлхийн банк Smart Government II PAD; АХБ TA-55211
- eID Mongolia: e-id.mn (лицензтэй CA; PIN1/PIN2, remote/batch signing, e-seal, qualified timestamp, OCSP/CRL, OAuth2/OIDC, mobile SDK, desktop + USB токен — 2026.06-ны байдлаар live); Smart-ID загварын лавлагаа: sk.ee (SK ID Solutions, Эстони)
- UN E-Government Survey 2024 (publicadministration.un.org)

*Тэмдэглэл: "hurdan.dgov.mn" нэрээр ажиллаж буй платформ судалгаагаар олдсонгүй — одоогийн албан ёсны домэйн khurdan.gov.mn (Төрийн цахим үйлчилгээний зохицуулалтын газар). Шинэ платформын нэршил гэж ойлгов.*
