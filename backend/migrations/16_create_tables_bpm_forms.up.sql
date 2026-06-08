-- bpm_forms: олон процесс дунд хуваалцаж ашиглах форм сан (Camunda-ийн "linked /
-- standalone form"-ийн эквивалент). Процессын userTask-ийн formKey нь
-- `gerege-forms:<formId>` хэлбэрээр энэ сангийн формыг лавлана; engine нь
-- openTask үед хамгийн сүүлийн schema-г эндээс уншина (latest-wins). RLS нь
-- bpm_process_definitions-тэй ижил (self / admin / service).
CREATE TABLE IF NOT EXISTS bpm_forms(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL DEFAULT '',
    -- form-js схем (JSON) — DynamicForm/form-js viewer шууд уншина.
    schema JSONB NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz
);

CREATE INDEX idx_bpm_forms_user ON bpm_forms (user_id, created_at DESC);

ALTER TABLE bpm_forms ENABLE ROW LEVEL SECURITY;
ALTER TABLE bpm_forms FORCE ROW LEVEL SECURITY;

CREATE POLICY bpm_forms_select ON bpm_forms
    FOR SELECT
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY bpm_forms_insert ON bpm_forms
    FOR INSERT
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY bpm_forms_update ON bpm_forms
    FOR UPDATE
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    )
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY bpm_forms_delete ON bpm_forms
    FOR DELETE
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );
