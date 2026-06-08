DELETE FROM role_permissions WHERE permission_key IN ('personal.view','manager.view');
DELETE FROM permissions WHERE key IN ('personal.view','manager.view');
