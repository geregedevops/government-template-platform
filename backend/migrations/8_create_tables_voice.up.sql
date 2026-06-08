-- Voice translation tables: MN<->EN speech translations and token-usage
-- metering. Same RLS model as the users/AI tables (migration 6 & 7): the app
-- sets app.user_id / app.user_role via set_config per transaction; a plain
-- user only touches rows they own, service/admin see everything, and a
-- missing identity denies all rows.

CREATE TABLE IF NOT EXISTS voice_translations(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_lang VARCHAR(8) NOT NULL CHECK (source_lang IN ('mn', 'en')),
    target_lang VARCHAR(8) NOT NULL CHECK (target_lang IN ('mn', 'en')),
    source_text TEXT NOT NULL DEFAULT '',
    translated_text TEXT NOT NULL DEFAULT '',
    model VARCHAR(64) NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_voice_translations_user ON voice_translations (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS voice_usage(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- nullable so a deleted translation does not erase its billing row.
    translation_id uuid REFERENCES voice_translations(id) ON DELETE SET NULL,
    model VARCHAR(64) NOT NULL,
    input_tokens integer NOT NULL DEFAULT 0,
    output_tokens integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_voice_usage_user ON voice_usage (user_id, created_at DESC);

-- ---------------------------------------------------------------------------
-- Row-Level Security (see migration 6 for the full rationale; FORCE is
-- required because the app connects as the table owner).
-- ---------------------------------------------------------------------------

ALTER TABLE voice_translations ENABLE ROW LEVEL SECURITY;
ALTER TABLE voice_translations FORCE ROW LEVEL SECURITY;

CREATE POLICY voice_translations_select ON voice_translations
    FOR SELECT
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY voice_translations_insert ON voice_translations
    FOR INSERT
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

-- Translations are immutable: no UPDATE policy on purpose (deny by default).
CREATE POLICY voice_translations_delete ON voice_translations
    FOR DELETE
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

ALTER TABLE voice_usage ENABLE ROW LEVEL SECURITY;
ALTER TABLE voice_usage FORCE ROW LEVEL SECURITY;

CREATE POLICY voice_usage_select ON voice_usage
    FOR SELECT
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY voice_usage_insert ON voice_usage
    FOR INSERT
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

-- Usage rows are append-only for non-admins: no UPDATE/DELETE policies
-- (deny by default) — billing/audit data must not be user-mutable.
