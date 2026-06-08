-- ROADMAP Үе 1: федерацийн цөм — platform registry (peers), гарын үсэгтэй
-- мессежийн outbox (durability) ба inbox (dedup/idempotency).
-- Бүгд admin/service-д л нээлттэй (федераци нь admin-удирдлагатай).

-- Гишүүн node-ууд (platform registry). jwks_url-аас гарын үсгийн нийтийн
-- түлхүүрийг (kid-ээр) татаж мессежийг шалгана.
CREATE TABLE IF NOT EXISTS fed_peers(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(120) UNIQUE NOT NULL,        -- node-ийн ялгах нэр (envelope iss/aud)
    name VARCHAR(200) NOT NULL DEFAULT '',
    org_id uuid REFERENCES organizations(id) ON DELETE SET NULL,
    base_url TEXT NOT NULL DEFAULT '',        -- /api/v1/fed/inbound энд нэмэгдэнэ
    jwks_url TEXT NOT NULL DEFAULT '',        -- нийтийн түлхүүрийн эх
    status VARCHAR(16) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'active', 'suspended')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz
);
CREATE INDEX idx_fed_peers_key ON fed_peers (key);

-- Гарах мессежийн дараалал (at-least-once + backoff retry + dead-letter).
CREATE TABLE IF NOT EXISTS fed_outbox(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    peer_id uuid NOT NULL REFERENCES fed_peers(id) ON DELETE CASCADE,
    typ VARCHAR(64) NOT NULL,
    envelope TEXT NOT NULL,                    -- гарын үсэгтэй JWS
    status VARCHAR(16) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'sent', 'dead')),
    attempts INT NOT NULL DEFAULT 0,
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    last_error TEXT NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_fed_outbox_due ON fed_outbox (status, next_attempt_at);

-- Хүлээн авсан мессежийн dedup (idempotency by jti) — давхар боловсруулахаас
-- сэргийлж "effectively-once" болгоно.
CREATE TABLE IF NOT EXISTS fed_inbox(
    jti uuid PRIMARY KEY,
    typ VARCHAR(64) NOT NULL,
    iss VARCHAR(120) NOT NULL,
    received_at timestamptz NOT NULL DEFAULT now()
);

-- RLS — бүгд admin/service-д л.
DO $$
DECLARE t text;
BEGIN
  FOREACH t IN ARRAY ARRAY['fed_peers','fed_outbox','fed_inbox'] LOOP
    EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
    EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
    EXECUTE format($f$CREATE POLICY %1$s_all ON %1$s FOR ALL
        USING (current_setting('app.user_role', true) IN ('service','admin'))
        WITH CHECK (current_setting('app.user_role', true) IN ('service','admin'))$f$, t);
  END LOOP;
END $$;

INSERT INTO permissions(key, label, category)
VALUES ('fed.manage', 'Федераци удирдах', 'Захиргаа')
ON CONFLICT (key) DO NOTHING;
