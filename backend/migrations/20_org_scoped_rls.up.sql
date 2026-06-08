-- ROADMAP Үе 0 / part 2: org-scoped RLS — байгууллагын модыг хандалтын хяналтад
-- холбоно. Хэрэглэгч зөвхөн ӨӨРИЙН байгууллагын ДЭД модыг хардаг/удирддаг
-- (root admin бүгдийг — бүх зүйл root-оос гаралтай). app.user_org GUC-ийг
-- (JWT-ийн OrgID claim → withRLS) уншина.
--
-- org_path_of нь SECURITY DEFINER — policy дотор organizations-ийг өөрийг нь
-- лавлахад RLS рекурс гарахаас сэргийлж, RLS-ийг тойрч path-ийг буцаана.
CREATE OR REPLACE FUNCTION org_path_of(p_org uuid) RETURNS ltree
LANGUAGE sql STABLE SECURITY DEFINER
SET search_path = public
AS $$
  SELECT path FROM organizations WHERE id = p_org
$$;

-- app.user_org хоосон (хуучин токен / service / org-гүй) бол org-хязгаар тавихгүй
-- (nullif ... IS NULL → бүх мөр) — backward compatible, эвдрэлгүй.
ALTER POLICY organizations_select ON organizations
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin', 'user')
        AND (
            nullif(current_setting('app.user_org', true), '') IS NULL
            OR path <@ org_path_of(nullif(current_setting('app.user_org', true), '')::uuid)
        )
    );

ALTER POLICY organizations_write ON organizations
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        AND (
            nullif(current_setting('app.user_org', true), '') IS NULL
            OR path <@ org_path_of(nullif(current_setting('app.user_org', true), '')::uuid)
        )
    )
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
        AND (
            nullif(current_setting('app.user_org', true), '') IS NULL
            OR path <@ org_path_of(nullif(current_setting('app.user_org', true), '')::uuid)
        )
    );
