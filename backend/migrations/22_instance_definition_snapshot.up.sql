-- ROADMAP Үе 0: процессын хувилбаржуулалт (execution safety). Гүйлт эхлэхэд
-- тухайн агшны .bpmn-ийг instance дотор snapshot болгож хадгална. Ингэснээр
-- процессын тодорхойлолтыг засахад АЖИЛЛАЖ БУЙ instance-ууд хөндөгдөхгүй
-- (өөрийн эхэлсэн хувилбараараа дуусна). Engine SubmitTask дээр одоогийн
-- тодорхойлолтыг дахин уншихын оронд snapshot-ийг ашиглана.
ALTER TABLE bpm_process_instances
    ADD COLUMN IF NOT EXISTS definition_snapshot text NOT NULL DEFAULT '';

-- Одоо ажиллаж буй instance-уудыг (snapshot хоосон) одоогийн тодорхойлолтоор
-- нөхнө (best-effort — хувилбаржуулалт хүчин төгөлдөр болохоос өмнөх гүйлтүүд).
UPDATE bpm_process_instances i
SET definition_snapshot = d.definition
FROM bpm_process_definitions d
WHERE d.id = i.definition_id AND i.definition_snapshot = '';
