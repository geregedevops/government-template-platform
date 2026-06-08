-- RLS-ийг migration 9-ийн (org-scope-гүй) хэлбэрт буцаана.
ALTER POLICY bpm_definitions_select ON bpm_process_definitions
    USING (current_setting('app.user_role', true) IN ('service', 'admin')
           OR user_id::text = current_setting('app.user_id', true));
ALTER POLICY bpm_definitions_insert ON bpm_process_definitions
    WITH CHECK (current_setting('app.user_role', true) IN ('service', 'admin')
               OR user_id::text = current_setting('app.user_id', true));
ALTER POLICY bpm_definitions_update ON bpm_process_definitions
    USING (current_setting('app.user_role', true) IN ('service', 'admin')
           OR user_id::text = current_setting('app.user_id', true))
    WITH CHECK (current_setting('app.user_role', true) IN ('service', 'admin')
               OR user_id::text = current_setting('app.user_id', true));
ALTER POLICY bpm_definitions_delete ON bpm_process_definitions
    USING (current_setting('app.user_role', true) IN ('service', 'admin')
           OR user_id::text = current_setting('app.user_id', true));

DROP INDEX IF EXISTS idx_bpmdef_org;
ALTER TABLE bpm_process_definitions DROP CONSTRAINT IF EXISTS fk_bpmdef_org;
ALTER TABLE bpm_process_definitions DROP COLUMN IF EXISTS org_id;
