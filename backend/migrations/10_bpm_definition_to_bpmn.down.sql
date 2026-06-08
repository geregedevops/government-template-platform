-- Буцаах: text → jsonb. Хадгалагдсан BPMN XML нь хүчинтэй JSON биш тул
-- агуулгыг хадгалж чадахгүй — хоосон объект болгоно (dev-ийн буцаалт).
ALTER TABLE bpm_process_definitions
    ALTER COLUMN definition TYPE jsonb USING '{}'::jsonb;
