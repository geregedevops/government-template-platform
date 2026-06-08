-- Хувийн (personal) болон Менежер (manager) системийн эрхүүдийг каталогт нэмнэ.
-- Эдгээр нь RBAC matrix-д автоматаар гарч ирнэ (category-аар бүлэглэгддэг).

INSERT INTO permissions(key, label, category) VALUES
    ('personal.view', 'Хувийн самбар',   'personal'),
    ('manager.view',  'Менежер самбар',   'manager')
ON CONFLICT (key) DO NOTHING;

-- Энгийн хэрэглэгч (role 2) одоо ч Хувийн системийг харна — өмнөх зан төлөвийг
-- хадгалахын тулд personal.view-г seed хийнэ. (admin нь Resolve дотор бүх
-- эрхийг автоматаар авдаг тул тусад нь seed шаардлагагүй.)
INSERT INTO role_permissions(role_id, permission_key) VALUES
    (2, 'personal.view')
ON CONFLICT DO NOTHING;
