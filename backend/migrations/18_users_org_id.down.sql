-- H6 trigger-ийг өмнөх (org_id-гүй) хэлбэрт буцаана.
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

DROP INDEX IF EXISTS idx_users_org_id;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_org;
ALTER TABLE users DROP COLUMN IF EXISTS org_id;
