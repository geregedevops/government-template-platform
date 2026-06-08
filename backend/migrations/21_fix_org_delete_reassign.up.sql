-- Засвар: users.org_id FK нь ON DELETE SET DEFAULT байсан ч migration 19
-- баганын DEFAULT-ийг хассан тул хэрэглэгчтэй (зөөлөн устгагдсаныг ч оруулаад)
-- байгууллагыг устгах үед алдаа гарч байв. Шийдэл: FK-г ON DELETE RESTRICT
-- болгож, BEFORE DELETE trigger-ээр тухайн байгууллагын хэрэглэгчдийг root руу
-- шилжүүлнэ (cascade subtree устгалд хүүхэд бүрд нь дуудагдана). GORM-аас
-- хамааралгүй (DB-түвшний), баганын default шаардахгүй тул auto-migrate churn-гүй.
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_org;
ALTER TABLE users
    ADD CONSTRAINT fk_users_org FOREIGN KEY (org_id)
    REFERENCES organizations(id) ON DELETE RESTRICT;

CREATE OR REPLACE FUNCTION reassign_users_on_org_delete() RETURNS trigger AS $$
BEGIN
  UPDATE users SET org_id = '00000000-0000-0000-0000-000000000001'
  WHERE org_id = OLD.id;
  RETURN OLD;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_reassign_users_before_org_delete ON organizations;
CREATE TRIGGER trg_reassign_users_before_org_delete
  BEFORE DELETE ON organizations
  FOR EACH ROW EXECUTE FUNCTION reassign_users_on_org_delete();
