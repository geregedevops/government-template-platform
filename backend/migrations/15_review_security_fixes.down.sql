-- 15-ийн урвуу.
DROP INDEX IF EXISTS idx_role_permissions_permission_key;
DROP INDEX IF EXISTS idx_bpm_events_user_id;
DROP INDEX IF EXISTS idx_bpm_tasks_user_id;
DROP INDEX IF EXISTS idx_voice_usage_translation_id;
DROP INDEX IF EXISTS idx_ai_usage_conversation_id;
DROP INDEX IF EXISTS idx_ai_messages_user_id;

ALTER POLICY roles_select ON roles USING (true);
ALTER POLICY permissions_select ON permissions USING (true);
ALTER POLICY role_permissions_select ON role_permissions USING (true);

DROP TRIGGER IF EXISTS trg_prevent_user_privilege_escalation ON users;
DROP FUNCTION IF EXISTS prevent_user_privilege_escalation();

ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_role_id;
ALTER TABLE users ALTER COLUMN role_id TYPE smallint;
