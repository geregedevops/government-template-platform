-- org-scope-гүй (migration 17-ийн) policy-уудад буцаана.
ALTER POLICY organizations_select ON organizations
    USING (current_setting('app.user_role', true) IN ('service', 'admin', 'user'));
ALTER POLICY organizations_write ON organizations
    USING (current_setting('app.user_role', true) IN ('service', 'admin'))
    WITH CHECK (current_setting('app.user_role', true) IN ('service', 'admin'));
DROP FUNCTION IF EXISTS org_path_of(uuid);
