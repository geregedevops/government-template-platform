// BPM (бизнес процесс)-ийн client талын типүүд. Процесс нь дэлхийн стандарт,
// зөөвөрлөх боломжтой нэг **BPMN 2.0 (.bpmn) файл** болж хадгалагдана. User
// task-ийн дэлгэцүүд (form-js схем) нь .bpmn дотроо Camunda-ийн стандарт
// zeebe:userTaskForm + zeebe:formDefinition extension element-ээр embed
// хийгдсэн байна — Camunda Modeler-ээр нээгдэх жинхэнэ .bpmn файл.

export type FormSchema = {
  type?: string;
  components?: unknown[];
  [key: string]: unknown;
};

/** serviceTask-ийн HTTP тохиргоо (zeebe:taskHeaders болж embed хийгдэнэ). */
export interface BpmService {
  method: string;
  url: string;
  body: string;
  resultVar: string;
}

/** Modeler доторх embed хийгдэх бүх тохиргоо. */
export interface BpmEmbeds {
  bpmn: string;
  forms: Record<string, FormSchema>; // userTaskId -> form-js схем
  services: Record<string, BpmService>; // serviceTaskId -> HTTP тохиргоо
  conditions: Record<string, string>; // sequenceFlowId -> нөхцөл
}

/** Backend-аас ирэх процессын DTO. bpmn нь цэвэр .bpmn XML. */
export interface BpmProcess {
  id: string;
  name: string;
  description: string;
  bpmn: string;
  status: 'draft' | 'published';
  version: number;
  created_at: string;
  updated_at: string | null;
}

/** Хуваалцсан форм (form library) — олон процесс дунд. */
export interface BpmForm {
  id: string;
  name: string;
  schema: FormSchema;
  created_at: string;
  updated_at: string | null;
}

/** Run-ийн нэг алхмын дэлгэц (user task-ийн form-js схем). */
export interface BpmTask {
  id: string;
  node_id: string;
  status: string;
  form: FormSchema;
}

export interface BpmInstance {
  id: string;
  definition_id: string;
  status: string;
  current_node_id: string;
  variables: Record<string, unknown>;
  created_at: string;
  completed_at: string | null;
}

export interface BpmRun {
  instance: BpmInstance;
  task: BpmTask | null;
  done: boolean;
}

/** Audit log-ийн нэг бичлэг (instance-ийн timeline). */
export interface BpmEvent {
  id: string;
  type: string;
  node_id: string;
  detail: string;
  created_at: string;
}

const BPMN_NS = 'http://www.omg.org/spec/BPMN/20100524/MODEL';
const ZEEBE_NS = 'http://camunda.org/schema/zeebe/1.0';
const FORM_KEY_PREFIX = 'camunda-forms:bpmn:';
/** Хуваалцсан (linked) формын formKey угтвар — backend-тэй ижил. */
export const SHARED_FORM_KEY_PREFIX = 'gerege-forms:';

/** forms map дахь хуваалцсан формын лавлагааны тэмдэг (inline schema биш). */
export function sharedFormRef(id: string): FormSchema {
  return { __sharedFormId: id } as unknown as FormSchema;
}
/** schema нь хуваалцсан формын лавлагаа эсэхийг шалгаж, ID-г буцаана. */
export function sharedFormId(schema: FormSchema | undefined): string | null {
  const v = schema as unknown as { __sharedFormId?: unknown } | undefined;
  return v && typeof v.__sharedFormId === 'string' ? v.__sharedFormId : null;
}

/** Хоосон form-js схем (шинэ user task-д). */
export function emptyFormSchema(): FormSchema {
  return { type: 'default', components: [] };
}

/**
 * User task-д форм байхгүй үед үүсгэх анхдагч маягт — гарчиг (процесс/таскийн
 * нэр) + дор хаяж нэг талбартай. Ингэснээр маягт хоосон харагдахгүй.
 */
export function defaultFormSchema(title: string): FormSchema {
  const rid = () => Math.random().toString(36).slice(2, 10);
  const safe = (title || 'Маягт').trim() || 'Маягт';
  return {
    type: 'default',
    id: `Form_${rid()}`,
    components: [
      { type: 'text', id: `Field_${rid()}`, text: `## ${safe}` },
      { type: 'textfield', id: `Field_${rid()}`, key: `field_${rid()}`, label: safe },
    ],
  };
}

/**
 * Шинэ процессын анхны BPMN 2.0 диаграмм — нэг start event-тэй (bpmn.io-ийн
 * стандарт хоосон загвар). bpmn-js нь байрлал (DI)-тай XML шаардана.
 */
export function emptyBpmn(): string {
  return `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL" xmlns:bpmndi="http://www.omg.org/spec/BPMN/20100524/DI" xmlns:dc="http://www.omg.org/spec/DD/20100524/DC" id="Definitions_1" targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:startEvent id="StartEvent_1" />
  </bpmn:process>
  <bpmndi:BPMNDiagram id="BPMNDiagram_1">
    <bpmndi:BPMNPlane id="BPMNPlane_1" bpmnElement="Process_1">
      <bpmndi:BPMNShape id="_BPMNShape_StartEvent_2" bpmnElement="StartEvent_1">
        <dc:Bounds x="173" y="102" width="36" height="36" />
      </bpmndi:BPMNShape>
    </bpmndi:BPMNPlane>
  </bpmndi:BPMNDiagram>
</bpmn:definitions>`;
}

/**
 * Хадгалагдсан .bpmn-ээс embed хийсэн тохиргоонуудыг (маягт, service HTTP,
 * gateway нөхцөл) ГАРГАЖ авч, тэдгээр extension-уудыг АРИЛГАСАН "цэвэр" BPMN-ийг
 * буцаана. bpmn-js нь zeebe moddle-гүйгээр ажилладаг тул эдгээрийг JS төлөвт
 * тусад нь авч, bpmn-js-д зөвхөн цэвэр BPMN өгнө. Browser-only (DOMParser).
 */
export function extractEmbeds(bpmn: string): BpmEmbeds {
  const forms: Record<string, FormSchema> = {};
  const services: Record<string, BpmService> = {};
  const conditions: Record<string, string> = {};
  if (typeof window === 'undefined' || !bpmn) return { bpmn, forms, services, conditions };
  const doc = new DOMParser().parseFromString(bpmn, 'application/xml');
  if (doc.getElementsByTagName('parsererror').length > 0) return { bpmn, forms, services, conditions };

  // --- маягт: userTaskForm + formDefinition ---
  const formById: Record<string, FormSchema> = {};
  for (const el of Array.from(doc.getElementsByTagNameNS(ZEEBE_NS, 'userTaskForm'))) {
    const id = el.getAttribute('id');
    if (id) {
      try {
        formById[id] = JSON.parse(el.textContent || '{}') as FormSchema;
      } catch {
        /* алгасах */
      }
    }
    el.parentNode?.removeChild(el);
  }
  for (const task of Array.from(doc.getElementsByTagNameNS(BPMN_NS, 'userTask'))) {
    const taskId = task.getAttribute('id');
    const fd = task.getElementsByTagNameNS(ZEEBE_NS, 'formDefinition')[0];
    if (taskId && fd) {
      const key = fd.getAttribute('formKey') || '';
      if (key.startsWith(SHARED_FORM_KEY_PREFIX)) {
        // Хуваалцсан (linked) форм — лавлагаа болгож тэмдэглэнэ (inline schema биш).
        const sid = key.slice(SHARED_FORM_KEY_PREFIX.length);
        if (sid) forms[taskId] = sharedFormRef(sid);
      } else {
        const fid = key.startsWith(FORM_KEY_PREFIX) ? key.slice(FORM_KEY_PREFIX.length) : '';
        if (fid && formById[fid]) forms[taskId] = formById[fid];
      }
      fd.parentNode?.removeChild(fd);
    }
  }

  // --- service task: zeebe:taskHeaders ---
  for (const task of Array.from(doc.getElementsByTagNameNS(BPMN_NS, 'serviceTask'))) {
    const taskId = task.getAttribute('id');
    const th = task.getElementsByTagNameNS(ZEEBE_NS, 'taskHeaders')[0];
    if (taskId && th) {
      const h: Record<string, string> = {};
      for (const header of Array.from(th.getElementsByTagNameNS(ZEEBE_NS, 'header'))) {
        const k = header.getAttribute('key');
        if (k) h[k] = header.getAttribute('value') || '';
      }
      services[taskId] = {
        method: h.method || 'GET',
        url: h.url || '',
        body: h.body || '',
        resultVar: h.resultVariable || '',
      };
      th.parentNode?.removeChild(th);
    }
  }

  // --- gateway нөхцөл: sequenceFlow > conditionExpression ---
  for (const flow of Array.from(doc.getElementsByTagNameNS(BPMN_NS, 'sequenceFlow'))) {
    const flowId = flow.getAttribute('id');
    const ce = flow.getElementsByTagNameNS(BPMN_NS, 'conditionExpression')[0];
    if (flowId && ce) {
      conditions[flowId] = cleanCondition(ce.textContent || '');
      ce.parentNode?.removeChild(ce);
    }
  }

  return { bpmn: new XMLSerializer().serializeToString(doc), forms, services, conditions };
}

/**
 * Цэвэр BPMN дээр бүх тохиргоог Camunda стандарт extension-уудаар EMBED хийж,
 * нэг бүрэн зөөвөрлөх .bpmn файл болгоно: маягт (zeebe:userTaskForm +
 * zeebe:formDefinition), service HTTP (zeebe:taskHeaders), gateway нөхцөл
 * (bpmn:conditionExpression). Browser-only.
 */
export function embedAll(
  cleanBpmn: string,
  forms: Record<string, FormSchema>,
  services: Record<string, BpmService>,
  conditions: Record<string, string>,
): string {
  if (typeof window === 'undefined') return cleanBpmn;
  const doc = new DOMParser().parseFromString(cleanBpmn, 'application/xml');
  if (doc.getElementsByTagName('parsererror').length > 0) return cleanBpmn;

  doc.documentElement.setAttribute('xmlns:zeebe', ZEEBE_NS);
  const process = doc.getElementsByTagNameNS(BPMN_NS, 'process')[0];
  if (!process) return cleanBpmn;

  // маягт
  const procExt = ensureExtensionElements(doc, process);
  for (const [taskId, schema] of Object.entries(forms)) {
    if (!findById(doc, BPMN_NS, 'userTask', taskId)) continue;
    const sharedId = sharedFormId(schema);
    if (sharedId) {
      // Хуваалцсан (linked) форм — зөвхөн лавлагаа (formKey) бичнэ; embedded
      // userTaskForm БИЧИХГҮЙ (engine нь openTask үед сангаас resolve хийнэ).
      const task = findById(doc, BPMN_NS, 'userTask', taskId);
      const taskExt = ensureExtensionElements(doc, task as Element);
      const fd = doc.createElementNS(ZEEBE_NS, 'zeebe:formDefinition');
      fd.setAttribute('formKey', SHARED_FORM_KEY_PREFIX + sharedId);
      taskExt.appendChild(fd);
      continue;
    }
    if (!schema || !Array.isArray(schema.components) || schema.components.length === 0) continue;
    const formId = `UserTaskForm_${taskId}`;
    const formEl = doc.createElementNS(ZEEBE_NS, 'zeebe:userTaskForm');
    formEl.setAttribute('id', formId);
    formEl.textContent = JSON.stringify(schema);
    procExt.appendChild(formEl);
    const task = findById(doc, BPMN_NS, 'userTask', taskId);
    const taskExt = ensureExtensionElements(doc, task as Element);
    const fd = doc.createElementNS(ZEEBE_NS, 'zeebe:formDefinition');
    fd.setAttribute('formKey', FORM_KEY_PREFIX + formId);
    taskExt.appendChild(fd);
  }

  // service HTTP
  for (const [taskId, cfg] of Object.entries(services)) {
    if (!cfg || !cfg.url.trim()) continue;
    const task = findById(doc, BPMN_NS, 'serviceTask', taskId);
    if (!task) continue;
    const ext = ensureExtensionElements(doc, task);
    const th = doc.createElementNS(ZEEBE_NS, 'zeebe:taskHeaders');
    addHeader(doc, th, 'method', cfg.method || 'GET');
    addHeader(doc, th, 'url', cfg.url.trim());
    if (cfg.body.trim()) addHeader(doc, th, 'body', cfg.body);
    if (cfg.resultVar.trim()) addHeader(doc, th, 'resultVariable', cfg.resultVar.trim());
    ext.appendChild(th);
  }

  // gateway нөхцөл
  for (const [flowId, cond] of Object.entries(conditions)) {
    if (!cond || !cond.trim()) continue;
    const flowEl = findById(doc, BPMN_NS, 'sequenceFlow', flowId);
    if (!flowEl) continue;
    const ce = doc.createElementNS(BPMN_NS, 'bpmn:conditionExpression');
    ce.textContent = '${' + cond.trim() + '}';
    flowEl.appendChild(ce);
  }

  return new XMLSerializer().serializeToString(doc);
}

function addHeader(doc: Document, parent: Element, key: string, value: string): void {
  const h = doc.createElementNS(ZEEBE_NS, 'zeebe:header');
  h.setAttribute('key', key);
  h.setAttribute('value', value);
  parent.appendChild(h);
}

// cleanCondition нь `${…}` боодол эсвэл эхний `=`-ийг хасч, доторх илэрхийллийг
// буцаана (backend-ийн cleanCondition-той ижил).
function cleanCondition(expr: string): string {
  let e = (expr || '').trim();
  if (e.startsWith('${') && e.endsWith('}')) e = e.slice(2, -1).trim();
  else if (e.startsWith('=')) e = e.slice(1).trim();
  return e;
}

function ensureExtensionElements(doc: Document, parent: Element): Element {
  for (const child of Array.from(parent.childNodes)) {
    if (
      child.nodeType === 1 &&
      (child as Element).namespaceURI === BPMN_NS &&
      (child as Element).localName === 'extensionElements'
    ) {
      return child as Element;
    }
  }
  const ext = doc.createElementNS(BPMN_NS, 'bpmn:extensionElements');
  parent.insertBefore(ext, parent.firstChild);
  return ext;
}

function findById(doc: Document, ns: string, local: string, id: string): Element | null {
  for (const el of Array.from(doc.getElementsByTagNameNS(ns, local))) {
    if (el.getAttribute('id') === id) return el;
  }
  return null;
}
