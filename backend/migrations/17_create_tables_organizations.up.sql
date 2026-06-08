-- Байгууллагын мод (organization hierarchy) — ROADMAP Үе 0 / P0.
-- Федератив hierarchy-ийн суурь: root → ministry → agency → soe. Материалжсан
-- зам нь ltree (`path`) — дэд модны асуулга `path <@ <scope>` (хойч үе) хурдан.
-- Ирээдүйд resource бүрд org_id + org-scoped RLS нэмэх суурь (одоо зөвхөн
-- registry; одоо байгаа хүснэгтүүдийг хөндөхгүй — additive).
CREATE EXTENSION IF NOT EXISTS ltree;

CREATE TABLE IF NOT EXISTS organizations(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    parent_id uuid REFERENCES organizations(id) ON DELETE CASCADE,
    -- Материалжсан зам: ltree label бүр `o<uuid_underscored>` (цифрээр эхлэхгүй,
    -- хүчинтэй ltree label). Жнь: o0000..001.o1111..222
    path ltree NOT NULL,
    name VARCHAR(200) NOT NULL DEFAULT '',
    kind VARCHAR(16) NOT NULL DEFAULT 'agency'
        CHECK (kind IN ('root', 'ministry', 'agency', 'soe')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz
);

CREATE INDEX idx_org_path ON organizations USING gist (path);
CREATE INDEX idx_org_parent ON organizations (parent_id);

ALTER TABLE organizations ENABLE ROW LEVEL SECURITY;
ALTER TABLE organizations FORCE ROW LEVEL SECURITY;

-- Унших: нэвтэрсэн дурын үүрэг (мод харагдана). Бичих: admin/service.
CREATE POLICY organizations_select ON organizations
    FOR SELECT
    USING (current_setting('app.user_role', true) IN ('service', 'admin', 'user'));
CREATE POLICY organizations_write ON organizations
    FOR ALL
    USING (current_setting('app.user_role', true) IN ('service', 'admin'))
    WITH CHECK (current_setting('app.user_role', true) IN ('service', 'admin'));

-- Үндэс байгууллага (root) — бүх hierarchy-ийн дээд.
INSERT INTO organizations(id, parent_id, path, name, kind)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    NULL,
    'o00000000_0000_0000_0000_000000000001',
    'Gerege',
    'root'
)
ON CONFLICT (id) DO NOTHING;

-- Эрхийн каталогт org.manage нэмнэ (admin автоматаар авна — ListPermissions).
INSERT INTO permissions(key, label, category)
VALUES ('org.manage', 'Байгууллага удирдах', 'Захиргаа')
ON CONFLICT (key) DO NOTHING;
