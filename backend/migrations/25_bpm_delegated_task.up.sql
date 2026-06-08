-- ROADMAP Үе 1: delegatedTask — нэг node-ийн процессын алхмыг өөр node (peer)
-- гүйцэтгэнэ. Гүйлтийн корреляци + 'waiting' (callback хүлээж буй) төлөв.
ALTER TABLE bpm_process_instances
    ADD COLUMN IF NOT EXISTS parent_instance_id uuid,
    ADD COLUMN IF NOT EXISTS origin_peer varchar(120) NOT NULL DEFAULT '';

-- status-д 'waiting' нэмнэ (delegatedTask дээр callback хүлээж буй instance).
ALTER TABLE bpm_process_instances DROP CONSTRAINT IF EXISTS bpm_process_instances_status_check;
ALTER TABLE bpm_process_instances ADD CONSTRAINT bpm_process_instances_status_check
    CHECK (status IN ('running', 'waiting', 'completed', 'cancelled', 'failed'));
