-- BPM (Business Process Management) tables: process definitions, running
-- instances, and user-task work items.
--
-- Same RLS model as the users / ai / voice tables (migrations 6/7/8): the app
-- sets app.user_id / app.user_role via SET LOCAL per transaction; a plain user
-- can only touch rows they own, service/admin see everything, and a missing
-- identity denies all rows (current_setting(..., true) -> NULL -> not true).
--
-- The process graph (nodes/edges/forms/api-config) is stored as a single JSONB
-- `definition` document — the canonical, AI-generatable, BPMN-inspired format
-- the React Flow modeler edits and the dynamic form renderer reads.

CREATE TABLE IF NOT EXISTS bpm_process_definitions(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    -- The whole process graph: { "nodes": [...], "edges": [...] }.
    definition JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(16) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published')),
    -- version is kept for forward-compatibility with a future "publish new
    -- version" flow; the foundation slice edits a single row in place (v1).
    version INTEGER NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz
);

CREATE INDEX idx_bpm_definitions_user ON bpm_process_definitions (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS bpm_process_instances(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    definition_id uuid NOT NULL REFERENCES bpm_process_definitions(id) ON DELETE CASCADE,
    -- user_id duplicates the definition owner on purpose: RLS can check
    -- ownership without a JOIN, and per-user audits stay cheap.
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(16) NOT NULL DEFAULT 'running'
        CHECK (status IN ('running', 'completed', 'cancelled', 'failed')),
    -- current_node_id is the BPMN-style token position (which node the run is at).
    current_node_id VARCHAR(120) NOT NULL DEFAULT '',
    -- Collected process variables (form submissions merged in): { "key": value }.
    variables JSONB NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz,
    completed_at timestamptz
);

CREATE INDEX idx_bpm_instances_definition ON bpm_process_instances (definition_id, created_at DESC);
CREATE INDEX idx_bpm_instances_user ON bpm_process_instances (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS bpm_tasks(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id uuid NOT NULL REFERENCES bpm_process_instances(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- node_id references the bpm form node this work item was opened for.
    node_id VARCHAR(120) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'completed')),
    -- Snapshot of the form spec at task-open time (so later edits to the
    -- definition don't change an in-flight task's rendered screen).
    form JSONB NOT NULL DEFAULT '{}',
    -- The user's submitted answers; NULL until the task is completed.
    submission JSONB,
    created_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz
);

CREATE INDEX idx_bpm_tasks_instance ON bpm_tasks (instance_id, status);

-- ---------------------------------------------------------------------------
-- Row-Level Security (see migration 6 for the full rationale; FORCE is
-- required because the app connects as the table owner).
-- ---------------------------------------------------------------------------

ALTER TABLE bpm_process_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE bpm_process_definitions FORCE ROW LEVEL SECURITY;

CREATE POLICY bpm_definitions_select ON bpm_process_definitions
    FOR SELECT
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY bpm_definitions_insert ON bpm_process_definitions
    FOR INSERT
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY bpm_definitions_update ON bpm_process_definitions
    FOR UPDATE
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    )
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY bpm_definitions_delete ON bpm_process_definitions
    FOR DELETE
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

ALTER TABLE bpm_process_instances ENABLE ROW LEVEL SECURITY;
ALTER TABLE bpm_process_instances FORCE ROW LEVEL SECURITY;

CREATE POLICY bpm_instances_select ON bpm_process_instances
    FOR SELECT
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY bpm_instances_insert ON bpm_process_instances
    FOR INSERT
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY bpm_instances_update ON bpm_process_instances
    FOR UPDATE
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    )
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

ALTER TABLE bpm_tasks ENABLE ROW LEVEL SECURITY;
ALTER TABLE bpm_tasks FORCE ROW LEVEL SECURITY;

CREATE POLICY bpm_tasks_select ON bpm_tasks
    FOR SELECT
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

CREATE POLICY bpm_tasks_insert ON bpm_tasks
    FOR INSERT
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );

-- Tasks are updated once (open -> completed with a submission); same
-- ownership check guards the UPDATE.
CREATE POLICY bpm_tasks_update ON bpm_tasks
    FOR UPDATE
    USING (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    )
    WITH CHECK (
        current_setting('app.user_role', true) IN ('service', 'admin')
        OR user_id::text = current_setting('app.user_id', true)
    );
