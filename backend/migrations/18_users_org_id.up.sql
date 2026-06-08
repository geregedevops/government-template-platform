-- ROADMAP Үе 0: хэрэглэгчийг байгууллагын модонд холбоно (users.org_id).
-- Энэ нь org-scoped эрх/RLS-ийн дата суурь (дараагийн increment-д JWT claim +
-- app.org_scope-оор enforce хийнэ). Одоо additive: бүх одоо байгаа хэрэглэгч
-- root байгууллагад хуваарилагдана.
ALTER TABLE users ADD COLUMN IF NOT EXISTS org_id uuid NOT NULL
    DEFAULT '00000000-0000-0000-0000-000000000001';

-- Байгууллага устгахад түүний хэрэглэгчид root руу шилжинэ (орхигдохгүй,
-- org устгалыг блоклохгүй).
ALTER TABLE users
    ADD CONSTRAINT fk_users_org FOREIGN KEY (org_id)
    REFERENCES organizations(id) ON DELETE SET DEFAULT;

CREATE INDEX IF NOT EXISTS idx_users_org_id ON users (org_id);

-- H6 trigger-ийг өргөтгөж, энгийн хэрэглэгч өөрийн org_id-г (мөн role_id/active)
-- өөрчлөхийг хориглоно (privilege escalation хамгаалалт).
CREATE OR REPLACE FUNCTION prevent_user_privilege_escalation() RETURNS trigger AS $$
BEGIN
  IF coalesce(current_setting('app.user_role', true), '') NOT IN ('', 'admin', 'service') THEN
    IF NEW.role_id IS DISTINCT FROM OLD.role_id
       OR NEW.active IS DISTINCT FROM OLD.active
       OR NEW.org_id IS DISTINCT FROM OLD.org_id THEN
      RAISE EXCEPTION 'permission denied: cannot modify role_id, active, or org_id';
    END IF;
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
