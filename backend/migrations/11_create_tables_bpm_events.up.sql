-- BPM audit log: нэг гүйлтийн (instance) төлөв өөрчлөлт бүрийн append-only
-- бүртгэл. ai_usage-тэй ижил RLS загвар: эзэмшигч/admin/service л харна, бичих
-- эрх ч мөн хязгаарлагдана; UPDATE/DELETE бодлогогүй (default-deny) тул audit
-- мөрийг хэрэглэгч өөрчилж/устгаж чадахгүй.

CREATE TABLE IF NOT EXISTS bpm_events(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id uuid NOT NULL REFERENCES bpm_process_instances(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(32) NOT NULL,
    node_id VARCHAR(120) NOT NULL DEFAULT '',
    detail TEXT NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_bpm_events_instance ON bpm_events (instance_id, created_at);

ALTER TABLE bpm_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE bpm_events FORCE ROW LEVEL SECURITY;

CREATE POLICY bpm_events_select ON bpm_events
    FOR SELECT
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY bpm_events_insert ON bpm_events
    FOR INSERT
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

-- Audit мөрүүд append-only: UPDATE/DELETE бодлогогүй (default-deny).
