# Кодын Review — Gerege Template AI BPM

Огноо: 2026-06-08 · Хамрах хүрээ: backend (Go), frontend (Next.js/TS), migration/RLS/infra

## Засагдсан төлөв (2026-06-08, deploy хийгдсэн)

**ЗАСАГДСАН ✅ (batch 1):** C1, C2, H1, H2, H3, H4, H5, H6, M3, M4, M7, M10, M13,
M14, M15, M16.

**ЗАСАГДСАН ✅ (batch 2):** H7 (бүх auth форм + Profile/Settings/ChangePassword/
SignOut/PasswordField i18n mn/en), M6 (voice mountedRef StrictMode), M9 (admin
page server-side perm), M11 (Down schema_migrations), M12 (advisory lock pooled
conn), + LOW glyph (login.verifyWithCode, forgot back-link). Build/test ногоон,
live.

**ЗАСАГДСАН ✅ (batch 3):** M5 (хэл/загвар propagation — shared useSyncExternalStore
store), M8 (chat stream AbortController + EOF flush), + LOW и-мэйл branding
(Go Rest → Gerege Template AI).

**ЗАСАГДСАН ✅ (batch 4):** M1 (role өөрчлөлт → token cutoff: TokenRevoker port,
Redis cutoff, refresh role_id дахин уншина — live баталгаажсан), M2 (lockout-ийг
(email, IP)-ээр + global per-email тоологч; client IP-г nginx X-Real-IP-ээс
[spoof-баталгаат] BFF→backend дамжуулна — security review-ийн HIGH олдворыг зассан;
live баталгаажсан), + LOW: access log → бүтэцлэгдсэн logger, EN огноо формат,
TTS onerror, sign-out toast, OTP email branding, email VARCHAR (live аль хэдийн
text).

**ҮЛДСЭН (зөвхөн жижиг cosmetic/infra — шаардлагатай бол):** UserMenu/AnonThemeToggle
inline lang===en, BPM cfg `<label htmlFor>` a11y, topbar search wiring,
activation regex (machine-readable код), schema_migrations RLS, in-memory
rate-limiter per-instance (Redis-backed болгох), compose api/web healthcheck
(image-д curl/wget байгаа эсэхийг шалгаад нэмэх).

---

## Ерөнхий дүгнэлт

Энэ кодын бааз нь аюулгүй байдлын хувьд **гайхалтай сайн суурьтай**: JWT-ийн алгоритмыг HS256-аар тогтоосон (alg-confusion халдлагаас хамгаалсан), RLS-ийг fail-closed горимоор бүх feature хүснэгт дээр хэрэгжүүлсэн, refresh token нэг удаагийн (atomic `GETDEL`), login-ийн хариуг тэгшитгэж timing leak-ийг хаасан, SSRF-ээс хамгаалсан connector-той, cookie-нууд httpOnly+Secure+SameSite=Strict, per-request nonce CSP, бүх mutating BFF route дээр `checkOrigin`. TypeScript-д `any` огт байхгүй.

Гэхдээ **deploy-д саад болохуйц хэд хэдэн ноцтой алдаа**, мөн **frontend дээр ажиллахгүй болгож буй логик алдаанууд**, төслийн өөрийн дүрэм зөрчсөн (i18n, икон) асуудлууд бий. Доорх жагсаалт ноцтойгоос нь эрэмбэлэв. Гол дүгнэлтүүдийг кодоос шууд шалгаж баталгаажуулсан.

---

## CRITICAL — нэн даруй засах

### C1. Migration-ийн файлууд лексикографиар эрэмблэгддэг — шинэ суулгац бүрэн эвдэрсэн
`backend/internal/datasources/migration/migration.go:191-198` (`sort.Strings`), `backend/migrations/*` (нэрс нь `1_`…`14_` — тэгээр нөхөөгүй)

Файлын нэрсийг **тэмдэгт мөрөөр** эрэмблэдэг тул бодит Up дараалал нь `10, 11, 12, 13, 14, 1, 2, …, 9` болдог. Цоо шинэ DB дээр хамгийн түрүүнд `10_bpm_definition_to_bpmn.up.sql` ажиллаж, migration 9 хүснэгтийг үүсгэхээс өмнө `ALTER TABLE bpm_process_definitions` хийх гээд **бүх Up унана**. Өөрөөр хэлбэл `docker compose up` нь шинэ deploy бүрд эвдэрнэ. Down мөн адил эсрэг асуудалтай (`9_…down.sql` нь `bpm_events`-ийн FK байсаар байхад `bpm_process_instances`-г DROP хийхийг оролдоно).

Засвар: файлуудыг тэгээр нөх (`01_…`, `02_…`, … `14_…`), эсвэл `listFiles`-д тоон эрэмбэ хий:
```go
sort.Slice(files, func(i, j int) bool { return migNum(files[i]) < migNum(files[j]) })
```

### C2. Seed нь таамаглахад хялбар нууц үгтэй admin данс үүсгэдэг, орчны хамгаалалтгүй
`backend/cmd/seed/seeders/seeder.users.go:18-30`, `backend/cmd/seed/main.go`

`patrick@gmail.com`-г `Active: true, RoleId: 1` (admin) болгож, нууц үг нь `"12345"`-ийн bcrypt hash. `cmd/seed/main.go`-д `ENVIRONMENT` шалгалт огт байхгүй тул `make seed`-ийг production DB рүү ажиллуулбал **таамаглахад хялбар нууц үгтэй backdoor admin** суулгана.

Засвар: `main()`-г production-д зогсоо (`if config.Environment == production { log.Fatal(...) }`), бодит admin seed-ийг сул нууц үгтэй бүү ачаалл.

---

## HIGH — яаралтай засах

### H1. AppShell нь admin биш бүх хэрэглэгчид анхны render дээр TypeError өгч унадаг
`frontend/src/components/AppShell.tsx:130-157`

Анхны render дээр `perms === null` тул `canSee()` бүх perm-тэй item-д `false` буцааж, `systems = []` болж, `activeSystem = systems.find(...) ?? systems[0]` нь `undefined`, улмаар 157-р мөрийн `useState(activeSystem.key)` нь *"Cannot read properties of undefined"* алдаа өгнө. Үүний улмаас **admin биш бүх роль дээр `/user/*`, `/manager/*` хуудсууд эвдэрнэ** (admin нь `isAdmin` short-circuit-аар зугтдаг).

Засвар: `?? systems[systems.length-1]`-ээр fallback хийж, `systems.length === 0` үед skeleton render хий. Хамгийн зөв нь server дээр аль хэдийн татсан `perms`-ийг prop-оор дамжуул.

### H2. RunClient нь BPM процессыг давхар эхлүүлдэг (давхар instance үүсгэнэ)
`frontend/src/components/bpm/RunClient.tsx:33-53`

`start` нь `useCallback([processId, T])`, харин `useEffect(() => void start(), [start])` нь `T`-ийн identity өөрчлөгдөх бүрд дахин ажилладаг. `T` нь mount-ийн дараа `usePreferences` localStorage-оос хэлийг sync хийхэд өөрчлөгддөг (хадгалсан хэл нь `en` бол), улмаар **хоёр дахь `POST /start` дуудагдаж давхар workflow instance үүснэ**. StrictMode dev-д бас давхарладаг.

Засвар: effect-ийг зөвхөн `processId`-аар түлхүүрлэ, `startedRef` хамгаалалт нэм.

### H3. RunClient нь start амжилтгүй болоход "Процесс дууслаа" гэж харуулдаг
`frontend/src/components/bpm/RunClient.tsx:87-102`

`const done = !run?.task` — `start()` амжилтгүй болоход `run === null` тул `done === true` болж, алдааны banner дээр "Процесс дууслаа" амжилтын карт зэрэг харагдана.

Засвар: `const done = !!run && !run.task;`, `run === null` үед алдаа/дахин оролдох төлөв харуул.

### H4. OTP enumeration — бүртгэлтэй и-мэйлийг таних боломж
`backend/internal/business/usecases/auth/auth.send_otp.go:55-81`, `auth.verify_otp.go:59-85`

`Login`/`ForgotPassword`-оос ялгаатай нь эдгээр route нь үл мэдэгдэх и-мэйлд `NotFound("email not found")`, идэвхтэйд `BadRequest("account already activated")` буцаадаг. Халдагч ямар и-мэйл бүртгэлтэй, идэвхтэй эсэхийг **enumerate хийх** боломжтой (rate-limit зөвхөн удаашруулна).

Засвар: данс байгаа эсэхээс үл хамааран ижил generic хариу буцаа (`ForgotPassword`-ийн decoy загвар).

### H5. Хуучин compose файл — superuser холболт (RLS бүрэн алгасагдана), default нууц үг, портууд нээлттэй
`backend/deploy/docker-compose.yml:17-43`

Root `docker-compose.yml` бүгдийг зөв хийсэн ч энэ хуучин файл эсрэгээрээ: `postgres/postgres` сул default, `5432`/`6379` портыг `0.0.0.0`-д нээсэн, Redis нууц үггүй, `migrate` service байхгүй тул API нь superuser-ээр холбогдож **6-13 migration-ийн бүх RLS policy-г алгасна**.

Засвар: энэ файлыг устга, эсвэл root compose-той ижил болго (non-superuser app role, портгүй, заавал нууц үг).

### H6. `users_update` RLS policy нь хэрэглэгчид өөрийн role_id-г өөрчлөх боломж олгодог (role escalation)
`backend/migrations/6_enable_rls_users.up.sql:50-59`

RLS зөвхөн мөрийн түвшинд. `users_update` policy нь `user` роль өөрийн мөрийг **аль багана**-г шинэчлэхийг хязгаарлахгүй — `role_id`, `active`, `password` оруулаад. App role-ээр SQL ажиллуулж чадсан халдагч `UPDATE users SET role_id=1 WHERE id=current_setting('app.user_id')` хийж admin болж чадна.

Засвар: багана түвшний grant (`REVOKE UPDATE … GRANT UPDATE (username,email,password,…)`), эсвэл `role_id`/`active` өөрчлөлтийг trigger-ээр блокло.

### H7. Auth-ийн бүх урсгал + хуваалцсан profile/settings component-ууд i18n дүрэм зөрчсөн (зөвхөн монголоор hardcode)
`frontend/src/app/{register,forgot-password,reset-password,verify-otp}/*Form.tsx`, `src/components/{ChangePasswordForm,ProfileSections,SettingsSections,SignOutButton,PasswordField}.tsx`

CLAUDE.md дүрмээр бүх UI текст `src/lib/i18n.ts`-д `mn`/`en` хоёулантай байх ёстой. Эдгээр component нь монголоор hardcode хийсэн, `useT`/`t()` дуудаагүй. Англи UI дээр эдгээр дэлгэц монголоор харагдана.

Засвар: бүх текстийг typed dictionary руу зөөж `useT()`-аар хэрэглэ.

---

## MEDIUM — засах нь зүйтэй

### Backend

- **M1. Role/permission өөрчлөлт token refresh болтол хэрэгжихгүй** — `pkg/jwt/jwt.go:135-157`, `middleware.rbac.go:36-48`. Admin хэрэглэгчийг бууруулсан ч token доторх хуучин `RoleID` нь дуустал (max 720h) хуучин эрх хадгална. Засвар: `UpdateRole` дээр token cutoff/version бичих.
- **M2. Per-email lockout нь targeted DoS боломжтой** — `auth.login.go:61-91`. Халдагч хохирогчийн и-мэйлд 10 буруу нууц үг оруулж 15 минут түгжиж чадна. Засвар: per-IP жинлэх эсвэл exponential backoff/CAPTCHA.
- **M3. BPM `MaxNodes` хамгаалалт нь dead code** — `bpm.impl.go:18-29`, `bpm.validate.go:241`. `MaxNodes` (200) хаана ч уншигдахгүй; зуу зуун service-task бүхий процесс нэг хүсэлтийг олон гадагш HTTP дуудлага болгоно. Засвар: `parseDefinition`-д node тоог хязгаарла.
- **M4. Динамик, орчуулаагүй мессеж** — `handlers/v1/auth/auth.send_otp.go:79`. `fmt.Sprintf("otp code has been send to %s", ...)` нь `catalogMN`-тай хэзээ ч таарахгүй (бас "send"→"sent" дүрмийн алдаа). Засвар: статик key ашиглаж и-мэйлийг `data`-д өг.

### Frontend

- **M5. Хэл солих нь тархдаггүй** — `src/lib/preferences.ts:51-89`, `useT.ts:14-18`. `usePreferences` нь component бүрд тусдаа `useState`; нэг газар хэл солиход бусад нь хуучин хэлээр үлддэг. Засвар: React context эсвэл `useSyncExternalStore`.
- **M6. `mountedRef` дахин `true` болдоггүй — voice бичлэг StrictMode-д ажиллахгүй** — `ChatClient.tsx:58-71`, `TranslateClient.tsx:43-64`. Засвар: effect-ийн эхэнд `mountedRef.current = true`.
- **M7. ProcessList delete нь хариуг үл тоодог** — `bpm/ProcessList.tsx:71-80`. Backend амжилтгүй болсон ч мөр UI-аас алга болоод дараагийн ачаалалт дээр эргэж гарна. Засвар: `body.ok` шалга.
- **M8. Chat streaming abort хийгддэггүй, сүүлийн SSE frame алдагдаж болзошгүй** — `ChatClient.tsx:99-156`. `AbortController` байхгүй; сүүлийн event `\n\n`-ээр төгсөөгүй бол `conversation_id` алдагдана. Засвар: AbortController нэм, EOF дээр buffer-ийг flush хий.
- **M9. Admin хуудсуудын server-side эрх шалгалт зөрүүтэй** — `admin/{chat,knowledge,translate,bpm}/page.tsx` нь зөвхөн session шалгадаг (users/roles нь `fetchMyPermissions` ашигладаг). Засвар: ижил perm шалгалт хэрэглэ.
- **M10. Path param-ийг encode хийлгүй backend URL-д залгадаг** — бүх `[id]` BFF route, ж: `api/users/[id]/route.ts:21`. `..%2Fadmin` гэх id нь proxy-г өөр endpoint руу чиглүүлж болзошгүй. Засвар: `encodeURIComponent(params.id)`.

### Migration / Infra

- **M11. `Down()` нь `schema_migrations`-ийг үл тоодог** — `migration.go:134-161`. Хэрэгжээгүй migration-уудыг ч буцаах гэж оролдоод унана. Засвар: applied биш бол алгас, `-steps` flag нэм.
- **M12. Advisory lock-ийг pooled `*sql.DB` дээр авдаг** — `migration.go:176-189`. Lock/unlock өөр өөр холболт дээр буудаг. Засвар: `r.db.Conn(ctx)`-аар нэг холболт pin хий.
- **M13. `users.role_id` нь `roles`-руу FK-гүй, төрөл зөрүүтэй** (`smallint` vs `integer`). Засвар: FK нэмж төрлийг нэгтгэ.
- **M14. FK / RLS-predicate багана дээр индекс дутуу** — `ai_messages.user_id`, `ai_usage.conversation_id`, `voice_usage.translation_id`, `bpm_tasks.user_id`, `bpm_events.user_id`, `role_permissions.permission_key`. Засвар: тус бүрд индекс үүсгэ.
- **M15. RBAC SELECT policy нь `USING (true)` — fail-open** — `13_create_tables_rbac.up.sql:56,63,70`. Бусад газрын deny-by-default зарчмыг зөрчдөг. Засвар: `current_setting('app.user_role', true) IN ('service','admin','user')`.
- **M16. initdb script нь fail-open superuser fallback** — `deploy/initdb/10-create-app-user.sh`. `APP_DB_USER` дутуу бол script exit 0 хийж app superuser-ээр чимээгүй ажиллана (RLS бүрэн алгасагдана). Бас нууц үгийг SQL-д шууд залгадаг (injection эрсдэл). Засвар: fail-closed (`exit 1`), `psql -v` ашигла.
- **M17. `ENVIRONMENT=development` deploy template-д** — `backend.env.example:6`. Dev горимд `/metrics`, `/swagger` нээлттэй болдог. Засвар: production горим + дотоод TLS exception дэмж.

---

## LOW — сайжруулалт

**Backend / Infra**

- Олон домэйн/auth алдааны мессеж `catalogMN`-д байхгүй (MN client англиар харна): `"username cannot be empty"`, `"email format is invalid"`, 429/413 body гэх мэт — `i18n.go`-д нэм.
- OTP и-мэйлд boilerplate branding үлдсэн — `mailer.otp.go:39-43` (`"Go Rest boilerplate"`, `"East Java, Indonesia"`).
- Access log нь zap-ийг тойрч `fmt.Printf`-аар ANSI өнгөтэй stdout руу бичдэг — `middleware.access_log.go:57-68`.
- In-memory rate limiter нь per-instance — олон replica дээр хязгаар нь `N×`. Redis-backed limiting санал болгоно.
- `email VARCHAR(50)` нь зарим хүчинтэй и-мэйлийг татгалзана — `VARCHAR(254)` болго.
- `schema_migrations`-д RLS байхгүй, `ALTER DEFAULT PRIVILEGES` нь ирээдүйн бүх хүснэгтийг app role-д бичих эрх өгдөг.
- Root compose-д `api`/`web` healthcheck байхгүй; Redis нууц үг command line-аар дамждаг (`docker inspect`-д харагдана).

**Frontend**

- Glyph икон зөрчил (CLAUDE.md STRICT): `forgot-password/ForgotPasswordForm.tsx:71` нь `←`-г back-link икон болгож ашигласан (ижил файлын 39-р мөр зөв `<ArrowLeft/>` хэрэглэсэн); `i18n.ts`-ийн `login.verifyWithCode` дахь trailing `→`. Засвар: lucide икон руу шилжүүл. *Эмодзи хаанаас ч олдсонгүй — энэ дүрэм биелсэн.*
- Англи UI дээр огноо монголоор харагдана — `DashboardView.tsx:39` (`formatDateMN`), `lib/format.ts`-д EN хувилбар алга.
- `UserMenu.tsx`, `AnonThemeToggle.tsx` нь inline `lang === 'en' ? …` ашигладаг (dictionary-г тойрдог).
- Sign-out амжилтгүй болоход чимээгүй — `signOut()`-ийг fire-and-forget дуудна, feedback алга.
- TTS "Listen" spinner гацаж болзошгүй — `ChatClient.tsx:275-281`, Audio-д `onerror` handler алга.
- Topbar хайлтын талбар нь wired хийгдээгүй dead UI — `AppShell.tsx:264-267`.
- BPM config panel-уудад `<label>` нь `htmlFor`-гүй (a11y) — `BpmModeler.tsx:346+`.
- Activation flow-г message regex-ээр илрүүлдэг (`LoginForm.tsx:64`) — backend machine-readable код буцаах нь зөв.

---

## Зөв хийгдсэн нь батлагдсан (арга хэмжээ шаардлагагүй)

- SQL injection алга: бүх raw SQL нь `?` parameter binding ашигладаг; RLS GUC нь parameterized `set_config(?, ?, true)` транзакц дотор.
- JWT нь HS256-г яг тогтоосон (alg-confusion-аас хамгаалсан); refresh token нэг удаагийн; нууц үг солиход өмнөх access token-ууд цуцлагддаг.
- FORCE ROW LEVEL SECURITY бүх RLS хүснэгт дээр байгаа; **root** compose нь `NOSUPERUSER NOBYPASSRLS` app role зөв үүсгэдэг.
- SSRF connector нь private/loopback/metadata IP-г блоклодог; voice/AI input-ууд хязгаарлагдсан; AI system prompt-д эмодзи хориглосон Formatting хэсэг байгаа.
- Frontend: token зөвхөн httpOnly+Secure+SameSite=Strict cookie-д; XSS байхгүй (`dangerouslySetInnerHTML` алга, react-markdown raw HTML-гүй); `safeNext()` open redirect блоклодог; бүх mutating route `checkOrigin`-тэй; `any` огт алга.

---

## Хамгийн түрүүнд хийх 6 зүйл

1. **C1** — migration файлуудыг тэгээр нөх (шинэ deploy ажиллахгүй байна).
2. **C2** — seed admin-ийг production-д блокло, сул нууц үгийг ав.
3. **H1** — AppShell-ийн undefined crash-ийг зас (admin биш бүх хэрэглэгч эвдэрсэн).
4. **H2/H3** — RunClient давхар instance + буруу "дууссан" төлөвийг зас.
5. **H5/H6** — хуучин compose-ийг устга, `users_update` багана түвшний эрхийг хязгаарла.
6. **H4** — OTP route-ийн email enumeration-ийг хаа.
