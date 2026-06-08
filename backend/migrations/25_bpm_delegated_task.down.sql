ALTER TABLE bpm_process_instances DROP CONSTRAINT IF EXISTS bpm_process_instances_status_check;
ALTER TABLE bpm_process_instances ADD CONSTRAINT bpm_process_instances_status_check
    CHECK (status IN ('running', 'completed', 'cancelled', 'failed'));
ALTER TABLE bpm_process_instances DROP COLUMN IF EXISTS origin_peer;
ALTER TABLE bpm_process_instances DROP COLUMN IF EXISTS parent_instance_id;
