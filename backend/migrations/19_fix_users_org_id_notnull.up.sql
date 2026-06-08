-- Засвар: migration 18-ийн дараах GORM auto-migrate нь users.org_id-ийн NOT NULL
-- + DEFAULT-ийг (type өөрчлөхийг оролдох явцдаа) хассан тул шинэ хэрэглэгчид
-- org_id = NULL болсон. Энд NULL-уудыг root болгож нөхөж, NOT NULL-ийг сэргээнэ.
-- DEFAULT-ийг САНААТАЙ хасна — апп (users.store.go) org_id-г тодорхой өгдөг тул
-- GORM-ийн загвар (type:uuid;not null, default-гүй) DB-тэй яг таарч, дахин
-- auto-migrate churn гарахгүй.
UPDATE users SET org_id = '00000000-0000-0000-0000-000000000001'
WHERE org_id IS NULL;

ALTER TABLE users ALTER COLUMN org_id DROP DEFAULT;
ALTER TABLE users ALTER COLUMN org_id SET NOT NULL;
