DELETE FROM permissions WHERE key = 'org.manage';
DROP POLICY IF EXISTS organizations_write ON organizations;
DROP POLICY IF EXISTS organizations_select ON organizations;
DROP TABLE IF EXISTS organizations;
-- ltree extension-ийг үлдээнэ (бусад зүйл хэрэглэж болзошгүй).
