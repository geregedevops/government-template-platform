# Gerege Template — Project Rules

## Golden rules (platform-wide)

- **Data entry is done ONLY by the interested/affected party — never by staff on
  their behalf.** Across every process and form on the platform, the person who
  has the stake in the data is the one who enters it. The applicant / citizen /
  customer fills in their own information and uploads their own materials;
  back-office / government staff (e.g. an officer at the National Auto Transport
  Center) **only verify, approve, reject, or route** — they do not re-type the
  applicant's data. (Mongolian: «Data entry-г ашиг сонирхолтой нь хүн л оруулна».)
  - When designing a BPM process: the first user task(s) belong to the affected
    party (submit everything required); staff tasks are verification/decision
    steps with select/checkbox/notes only — incomplete → return for correction,
    complete → continue. The reference process is **"Автомашины эзэмшигч солих"**.

## Icons & emoji (STRICT)

- **Use flat icons everywhere — never emoji.** In the frontend, all icons come from
  `lucide-react` (flat SVG). Do not use emoji (🍰, 🌐, 😊, ✅, ⚡, 📌 …) or glyph
  characters (`←`, `→`, `↔`) as standalone icons. For a back link use
  `<ArrowLeft size={14} strokeWidth={2} />`, not a `←` character.
  - Typographic arrows inside running text/labels (e.g. "Монгол → Англи") are fine;
    the rule targets icons, buttons, and decorative markers.
- **The AI chat assistant must not emit emoji.** This is enforced in the system prompt
  at `backend/internal/business/usecases/ai/ai.system_prompt.go` (the "Formatting"
  section). If chat responses start showing emoji again, check that prompt first.

## Conventions worth keeping in mind

- New features mirror the AI chat feature's Clean Architecture exactly (handlers ->
  usecases -> repositories; RLS on every table; `*.ai.go` / `*.voice.go` file naming;
  bilingual mn/en comments). See the voice feature for the most recent example.
- User-visible backend messages are English keys translated in
  `backend/internal/i18n/i18n.go` (`catalogMN`) — add a translation for every new message.
- Frontend strings live in `frontend/src/lib/i18n.ts` (typed `mn`/`en` dictionaries) —
  add both languages or the build fails.
- Secrets stay backend-only (never sent to the browser); BFF routes attach the Bearer
  token server-side and guard with `checkOrigin` on POST routes.
