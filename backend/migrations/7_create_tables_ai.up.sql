-- AI chat tables: conversations, messages, and token-usage metering.
--
-- Same RLS model as the users table (migration 6): the app sets
-- app.user_id / app.user_role via SET LOCAL per transaction; a plain user
-- can only touch rows they own, service/admin see everything, and a missing
-- identity denies all rows (current_setting(..., true) -> NULL -> not true).

CREATE TABLE IF NOT EXISTS ai_conversations(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(120) NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz
);

CREATE INDEX idx_ai_conversations_user ON ai_conversations (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS ai_messages(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id uuid NOT NULL REFERENCES ai_conversations(id) ON DELETE CASCADE,
    -- user_id duplicates the conversation owner on purpose: RLS policies can
    -- check ownership without a JOIN, and the column makes per-user audits cheap.
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(16) NOT NULL CHECK (role IN ('user', 'assistant')),
    content TEXT NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_ai_messages_conversation ON ai_messages (conversation_id, created_at);

CREATE TABLE IF NOT EXISTS ai_usage(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    conversation_id uuid REFERENCES ai_conversations(id) ON DELETE SET NULL,
    model VARCHAR(64) NOT NULL,
    input_tokens integer NOT NULL DEFAULT 0,
    output_tokens integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_ai_usage_user ON ai_usage (user_id, created_at DESC);

-- ---------------------------------------------------------------------------
-- Row-Level Security (see migration 6 for the full rationale; FORCE is
-- required because the app connects as the table owner).
-- ---------------------------------------------------------------------------

ALTER TABLE ai_conversations ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_conversations FORCE ROW LEVEL SECURITY;

CREATE POLICY ai_conversations_select ON ai_conversations
    FOR SELECT
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY ai_conversations_insert ON ai_conversations
    FOR INSERT
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY ai_conversations_update ON ai_conversations
    FOR UPDATE
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    )
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY ai_conversations_delete ON ai_conversations
    FOR DELETE
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

ALTER TABLE ai_messages ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_messages FORCE ROW LEVEL SECURITY;

CREATE POLICY ai_messages_select ON ai_messages
    FOR SELECT
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY ai_messages_insert ON ai_messages
    FOR INSERT
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

-- Messages are immutable: no UPDATE policy on purpose (deny by default).
CREATE POLICY ai_messages_delete ON ai_messages
    FOR DELETE
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

ALTER TABLE ai_usage ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_usage FORCE ROW LEVEL SECURITY;

CREATE POLICY ai_usage_select ON ai_usage
    FOR SELECT
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY ai_usage_insert ON ai_usage
    FOR INSERT
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

-- Usage rows are append-only for non-admins: no UPDATE/DELETE policies
-- (deny by default) — billing/audit data must not be user-mutable.
