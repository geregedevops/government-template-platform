DROP TRIGGER IF EXISTS trg_reassign_users_before_org_delete ON organizations;
DROP FUNCTION IF EXISTS reassign_users_on_org_delete();
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_org;
ALTER TABLE users
    ADD CONSTRAINT fk_users_org FOREIGN KEY (org_id)
    REFERENCES organizations(id) ON DELETE SET DEFAULT;
