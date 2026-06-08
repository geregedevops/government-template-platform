-- RBAC: динамик эрх (roles) + эрхийн каталог (permissions) + role↔permission
-- холбоос. users.role_id нь roles.id-г заана (одоо 1=admin, 2=user). Эдгээр нь
-- глобал тохиргооны хүснэгтүүд тул RLS-д: SELECT нь бүх нэвтэрсэн хэрэглэгчид
-- нээлттэй (хэрэглэгч өөрийн эрхээ тооцоолоход хэрэгтэй), бичих нь зөвхөн
-- admin/service.

CREATE TABLE IF NOT EXISTS roles(
    id serial PRIMARY KEY,
    key varchar(50) UNIQUE NOT NULL,
    name varchar(100) NOT NULL,
    description text NOT NULL DEFAULT '',
    is_system boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz
);

-- Одоо байгаа role_id 1/2-тэй тааруулж систем эрхүүдийг seed хийнэ.
INSERT INTO roles(id, key, name, description, is_system) VALUES
    (1, 'admin', 'Админ', 'Бүх эрхтэй системийн админ', true),
    (2, 'user',  'Хэрэглэгч', 'Энгийн хэрэглэгч', true)
ON CONFLICT (id) DO NOTHING;
-- serial-ийн дараагийн утгыг seed-ийн дараа зөв болгоно.
SELECT setval('roles_id_seq', GREATEST((SELECT MAX(id) FROM roles), 1));

CREATE TABLE IF NOT EXISTS permissions(
    key varchar(64) PRIMARY KEY,
    label varchar(120) NOT NULL,
    category varchar(64) NOT NULL DEFAULT ''
);

INSERT INTO permissions(key, label, category) VALUES
    ('dashboard.view',   'Хяналтын самбар',        'general'),
    ('settings.manage',  'Аюулгүй байдлын тохиргоо', 'general'),
    ('ai.chat',          'AI чат',                  'ai'),
    ('knowledge.manage', 'Мэдлэгийн сан',           'ai'),
    ('voice.translate',  'Дуу хоолойн орчуулга',    'ai'),
    ('bpm.manage',       'Бизнес процесс',          'ai'),
    ('users.manage',     'Хэрэглэгч удирдах',       'admin'),
    ('roles.manage',     'Эрх удирдах (RBAC)',      'admin')
ON CONFLICT (key) DO NOTHING;

CREATE TABLE IF NOT EXISTS role_permissions(
    role_id integer NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_key varchar(64) NOT NULL REFERENCES permissions(key) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_key)
);

-- Admin (role 1) бүх эрхтэй.
INSERT INTO role_permissions(role_id, permission_key)
    SELECT 1, key FROM permissions
ON CONFLICT DO NOTHING;

-- RLS: SELECT нээлттэй (эрх тооцоолох), бичих admin/service.
ALTER TABLE roles ENABLE ROW LEVEL SECURITY;
ALTER TABLE roles FORCE ROW LEVEL SECURITY;
CREATE POLICY roles_select ON roles FOR SELECT USING (true);
CREATE POLICY roles_write ON roles FOR ALL
    USING (current_setting('app.user_role', true) IN ('service', 'admin'))
    WITH CHECK (current_setting('app.user_role', true) IN ('service', 'admin'));

ALTER TABLE permissions ENABLE ROW LEVEL SECURITY;
ALTER TABLE permissions FORCE ROW LEVEL SECURITY;
CREATE POLICY permissions_select ON permissions FOR SELECT USING (true);
CREATE POLICY permissions_write ON permissions FOR ALL
    USING (current_setting('app.user_role', true) IN ('service', 'admin'))
    WITH CHECK (current_setting('app.user_role', true) IN ('service', 'admin'));

ALTER TABLE role_permissions ENABLE ROW LEVEL SECURITY;
ALTER TABLE role_permissions FORCE ROW LEVEL SECURITY;
CREATE POLICY role_permissions_select ON role_permissions FOR SELECT USING (true);
CREATE POLICY role_permissions_write ON role_permissions FOR ALL
    USING (current_setting('app.user_role', true) IN ('service', 'admin'))
    WITH CHECK (current_setting('app.user_role', true) IN ('service', 'admin'));
