-- Процессыг дэлхийн стандартаар хадгалах refactor: bpm_process_definitions.definition
-- нь одоо custom JSON ({bpmn,forms}) бус, цэвэр BPMN 2.0 XML (.bpmn) файл болж
-- хадгалагдана (маягтууд нь Camunda zeebe:userTaskForm extension element-ээр XML
-- дотроо embed хийгдсэн). XML нь JSON биш тул баганыг jsonb → text болгоно.
--
-- bpm_tasks.form (form-js схем) болон bpm_process_instances.variables нь жинхэнэ
-- JSON хэвээр тул jsonb-ээр үлдэнэ.

ALTER TABLE bpm_process_definitions
    ALTER COLUMN definition TYPE text USING definition::text;
