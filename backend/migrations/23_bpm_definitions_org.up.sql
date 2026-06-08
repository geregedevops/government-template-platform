-- ROADMAP Үе 0: процессын тодорхойлолтыг байгууллагын модонд холбоно.
-- Хэрэглэгч өөрийн процессоо, харин admin нь ӨӨРИЙН байгууллагын дэд модны бүх
-- процессыг хардаг/удирддаг (root admin бүгдийг). org_path_of (migration 20)
-- + app.user_org (JWT OrgID) ашиглана. definitions нь GORM auto-migrate-д
-- байхгүй тул баганын DEFAULT + ON DELETE SET DEFAULT найдвартай.
ALTER TABLE bpm_process_definitions ADD COLUMN IF NOT EXISTS org_id uuid NOT NULL
    DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE bpm_process_definitions
    ADD CONSTRAINT fk_bpmdef_org FOREIGN KEY (org_id)
    REFERENCES organizations(id) ON DELETE SET DEFAULT;
CREATE INDEX IF NOT EXISTS idx_bpmdef_org ON bpm_process_definitions (org_id);

-- RLS: service бүгд; эзэмшигч өөрийн; admin өөрийн org-subtree (scope хоосон/
-- хуучин токен бол бүгд — backward compatible).
ALTER POLICY bpm_definitions_select ON bpm_process_definitions
    USING (
        current_setting('app.user_role', true) = 'service'
        OR user_id::text = current_setting('app.user_id', true)
        OR (current_setting('app.user_role', true) = 'admin'
            AND (nullif(current_setting('app.user_org', true), '') IS NULL
                 OR org_path_of(org_id) <@ org_path_of(nullif(current_setting('app.user_org', true), '')::uuid)))
    );

ALTER POLICY bpm_definitions_insert ON bpm_process_definitions
    WITH CHECK (
        current_setting('app.user_role', true) = 'service'
        OR user_id::text = current_setting('app.user_id', true)
        OR (current_setting('app.user_role', true) = 'admin'
            AND (nullif(current_setting('app.user_org', true), '') IS NULL
                 OR org_path_of(org_id) <@ org_path_of(nullif(current_setting('app.user_org', true), '')::uuid)))
    );

ALTER POLICY bpm_definitions_update ON bpm_process_definitions
    USING (
        current_setting('app.user_role', true) = 'service'
        OR user_id::text = current_setting('app.user_id', true)
        OR (current_setting('app.user_role', true) = 'admin'
            AND (nullif(current_setting('app.user_org', true), '') IS NULL
                 OR org_path_of(org_id) <@ org_path_of(nullif(current_setting('app.user_org', true), '')::uuid)))
    )
    WITH CHECK (
        current_setting('app.user_role', true) = 'service'
        OR user_id::text = current_setting('app.user_id', true)
        OR (current_setting('app.user_role', true) = 'admin'
            AND (nullif(current_setting('app.user_org', true), '') IS NULL
                 OR org_path_of(org_id) <@ org_path_of(nullif(current_setting('app.user_org', true), '')::uuid)))
    );

ALTER POLICY bpm_definitions_delete ON bpm_process_definitions
    USING (
        current_setting('app.user_role', true) = 'service'
        OR user_id::text = current_setting('app.user_id', true)
        OR (current_setting('app.user_role', true) = 'admin'
            AND (nullif(current_setting('app.user_org', true), '') IS NULL
                 OR org_path_of(org_id) <@ org_path_of(nullif(current_setting('app.user_org', true), '')::uuid)))
    );
