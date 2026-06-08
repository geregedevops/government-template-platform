-- AI knowledge base: хэрэглэгчийн өөрийн оруулсан мэдлэгийн бичлэгүүд. AI чат нь
-- эдгээрийг system prompt-д шигтгэж, платформоос гадуурх (эсвэл тусгай) асуултад
-- хариулахад ашиглана. ai_conversations-тэй ижил RLS загвар (self/admin/service).

CREATE TABLE IF NOT EXISTS ai_knowledge(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL DEFAULT '',
    content TEXT NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz
);

CREATE INDEX idx_ai_knowledge_user ON ai_knowledge (user_id, created_at DESC);

ALTER TABLE ai_knowledge ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_knowledge FORCE ROW LEVEL SECURITY;

CREATE POLICY ai_knowledge_select ON ai_knowledge
    FOR SELECT
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY ai_knowledge_insert ON ai_knowledge
    FOR INSERT
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY ai_knowledge_update ON ai_knowledge
    FOR UPDATE
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    )
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY ai_knowledge_delete ON ai_knowledge
    FOR DELETE
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );
