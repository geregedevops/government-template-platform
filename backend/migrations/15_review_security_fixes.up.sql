-- Gerege Template AI v1.0 — REVIEW.md-ийн аюулгүй байдал/бүрэн бүтэн байдлын засварууд.
-- (H6) хэрэглэгч өөрийн role_id/active-ээ өөрчлөхөөс сэргийлэх trigger,
-- (M13) users.role_id → roles(id) FK + төрлийн уялдаа, (M15) RBAC SELECT
-- policy-г fail-closed болгох, (M14) FK/RLS багана дээрх индексүүд.

-- --- M13: users.role_id бүрэн бүтэн байдал ---------------------------------
-- Эхлээд устгагдсан role-руу заасан "өнчин" role_id-уудыг user (2) болгоно
-- (trigger үүсгэхээс ӨМНӨ — эс бөгөөс энэ UPDATE өөрөө trigger-т баригдана).
UPDATE users SET role_id = 2
WHERE role_id NOT IN (SELECT id FROM roles);

ALTER TABLE users ALTER COLUMN role_id TYPE integer;
ALTER TABLE users
  ADD CONSTRAINT fk_users_role_id FOREIGN KEY (role_id)
  REFERENCES roles(id) ON DELETE RESTRICT;

-- --- H6: privilege escalation-аас сэргийлэх trigger ------------------------
-- RLS нь зөвхөн МӨР-ийн түвшинд; багана хязгаарлахгүй тул app role-ээр SQL
-- ажиллуулж чадсан халдагч өөрийн role_id/active-ээ өөрчилж admin болж болзошгүй.
-- Энэ trigger нь admin/service (эсвэл superuser, role хоосон) бус контекстээс
-- role_id эсвэл active өөрчлөхийг хориглоно.
CREATE OR REPLACE FUNCTION prevent_user_privilege_escalation() RETURNS trigger AS $$
BEGIN
  IF coalesce(current_setting('app.user_role', true), '') NOT IN ('', 'admin', 'service') THEN
    IF NEW.role_id IS DISTINCT FROM OLD.role_id OR NEW.active IS DISTINCT FROM OLD.active THEN
      RAISE EXCEPTION 'permission denied: cannot modify role_id or active';
    END IF;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_prevent_user_privilege_escalation ON users;
CREATE TRIGGER trg_prevent_user_privilege_escalation
  BEFORE UPDATE ON users
  FOR EACH ROW EXECUTE FUNCTION prevent_user_privilege_escalation();

-- --- M15: RBAC SELECT policy-г fail-closed болгох --------------------------
-- USING (true) нь deny-by-default зарчмыг зөрчдөг. Зөвхөн танигдсан үүрэгт
-- зөвшөөрнө (нэвтрээгүй/role-гүй контекст хаагдана).
ALTER POLICY roles_select ON roles
  USING (current_setting('app.user_role', true) IN ('service', 'admin', 'user'));
ALTER POLICY permissions_select ON permissions
  USING (current_setting('app.user_role', true) IN ('service', 'admin', 'user'));
ALTER POLICY role_permissions_select ON role_permissions
  USING (current_setting('app.user_role', true) IN ('service', 'admin', 'user'));

-- --- M14: FK / RLS-predicate багана дээрх индексүүд ------------------------
CREATE INDEX IF NOT EXISTS idx_ai_messages_user_id ON ai_messages (user_id);
CREATE INDEX IF NOT EXISTS idx_ai_usage_conversation_id ON ai_usage (conversation_id);
CREATE INDEX IF NOT EXISTS idx_voice_usage_translation_id ON voice_usage (translation_id);
CREATE INDEX IF NOT EXISTS idx_bpm_tasks_user_id ON bpm_tasks (user_id);
CREATE INDEX IF NOT EXISTS idx_bpm_events_user_id ON bpm_events (user_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_key ON role_permissions (permission_key);
