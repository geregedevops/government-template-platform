"use client";

import 'bpmn-js/dist/assets/diagram-js.css';
import 'bpmn-js/dist/assets/bpmn-js.css';
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn.css';
import '@bpmn-io/form-js/dist/assets/form-js-editor.css';
import '@bpmn-io/form-js/dist/assets/properties-panel.css';

import { useEffect, useRef, useState, type MutableRefObject } from 'react';
import { useRouter } from 'next/navigation';
import BpmnModeler from 'bpmn-js/lib/Modeler';
import { FormEditor } from '@bpmn-io/form-js';
import {
  Save, ArrowLeft, Loader2, FileText, MousePointerClick, Webhook, GitBranch,
  Maximize2, Minimize2, Maximize, Plus, Minus,
} from 'lucide-react';
import { useT } from '@/lib/useT';
import { postJSON } from '@/lib/client';
import {
  emptyBpmn, defaultFormSchema, extractEmbeds, embedAll,
  sharedFormId, sharedFormRef,
  type FormSchema, type BpmService, type BpmForm,
} from '@/lib/bpm';

interface Props {
  initial: {
    id?: string;
    name: string;
    description: string;
    bpmn: string;
  };
}

// eventBus-ийн сонголтын эвентийн хэлбэр (бидний ашигладаг хэсэг). connection
// (sequenceFlow) элемент дээр source/target байдаг.
interface BpmnEl {
  id: string;
  type: string;
  source?: { type: string };
}
interface SelectionEvent {
  newSelection?: BpmnEl[];
}
interface EventBus {
  on(event: string, callback: (e: SelectionEvent) => void): void;
}

type Selected = { id: string; type: string; sourceType?: string };

function isGatewayFlow(sel: Selected | null): boolean {
  return (
    sel?.type === 'bpmn:SequenceFlow' &&
    (sel.sourceType === 'bpmn:ExclusiveGateway' || sel.sourceType === 'bpmn:InclusiveGateway')
  );
}

export default function BpmModeler({ initial }: Props) {
  const { T } = useT();
  const router = useRouter();

  const canvasRef = useRef<HTMLDivElement>(null);
  const modelerRef = useRef<BpmnModeler | null>(null);
  const formHostRef = useRef<HTMLDivElement>(null);
  const formEditorRef = useRef<FormEditor | null>(null);
  // Тохиргоонуудыг render-ийн чичиргээнээс ангид ref-д хадгална. Эхний утгыг
  // init effect дотор .bpmn-ээс задалж populate хийнэ.
  const formsRef = useRef<Record<string, FormSchema>>({}); // userTaskId -> form схем
  const servicesRef = useRef<Record<string, BpmService>>({}); // serviceTaskId -> HTTP
  const conditionsRef = useRef<Record<string, string>>({}); // sequenceFlowId -> нөхцөл
  const currentTaskRef = useRef<string | null>(null);

  const [name, setName] = useState(initial.name);
  const [selected, setSelected] = useState<Selected | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [fullscreen, setFullscreen] = useState(false);
  // Маягтын эх сурвалж: inline (энэ процессод зурах) | shared (хуваалцсан сангаас).
  const [formMode, setFormMode] = useState<'inline' | 'shared'>('inline');
  const [sharedSel, setSharedSel] = useState('');
  const [sharedForms, setSharedForms] = useState<BpmForm[]>([]);

  const selectedTaskId = selected?.type === 'bpmn:UserTask' ? selected.id : null;

  // --- zoom / fullscreen хяналт ---
  type CanvasSvc = { zoom: (v?: number | string) => number; resized: () => void };
  const fitView = () => {
    try {
      modelerRef.current?.get<CanvasSvc>('canvas').zoom('fit-viewport');
    } catch {
      /* modeler бэлэн биш */
    }
  };
  const zoomBy = (delta: number) => {
    try {
      const c = modelerRef.current?.get<CanvasSvc>('canvas');
      if (c) {
        const z = c.zoom();
        c.zoom(Math.max(0.2, Math.min(4, (typeof z === 'number' ? z : 1) + delta)));
      }
    } catch {
      /* modeler бэлэн биш */
    }
  };

  // Fullscreen солигдоход канвасын хэмжээ өөрчлөгдөнө — bpmn-js-д мэдэгдэж
  // дахин тааруулна.
  useEffect(() => {
    const t = setTimeout(() => {
      try {
        const c = modelerRef.current?.get<CanvasSvc>('canvas');
        c?.resized();
        c?.zoom('fit-viewport');
      } catch {
        /* бэлэн биш */
      }
    }, 60);
    return () => clearTimeout(t);
  }, [fullscreen]);

  // Fullscreen үед Escape дарж гарна.
  useEffect(() => {
    if (!fullscreen) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setFullscreen(false);
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [fullscreen]);

  // bpmn-js modeler-ийг нэг удаа үүсгэж, диаграммыг ачаална. .bpmn доторх embed
  // хийсэн маягтуудыг гаргаж formsRef-д хийгээд, bpmn-js-д ЦЭВЭР BPMN өгнө.
  useEffect(() => {
    if (!canvasRef.current) return;
    const { bpmn: cleanBpmn, forms, services, conditions } = extractEmbeds(initial.bpmn || emptyBpmn());
    formsRef.current = forms;
    servicesRef.current = services;
    conditionsRef.current = conditions;
    const modeler = new BpmnModeler({ container: canvasRef.current });
    modelerRef.current = modeler;
    void modeler.importXML(cleanBpmn).then(() => {
      try {
        modeler.get<{ zoom: (v?: number | string) => number }>('canvas').zoom('fit-viewport');
      } catch {
        /* алгасах */
      }
      // Форм байхгүй user task бүрд анхдагч маягт (гарчиг + 1 талбар) суулгана —
      // ингэснээр маягт хэзээ ч хоосон биш, хадгалахад embed хийгдэнэ.
      try {
        type RegEl = { id: string; type: string; businessObject?: { name?: string } };
        const reg = modeler.get<{ filter: (fn: (el: RegEl) => boolean) => RegEl[] }>('elementRegistry');
        reg.filter((el) => el.type === 'bpmn:UserTask').forEach((el) => {
          const existing = formsRef.current[el.id];
          if (!existing || !Array.isArray(existing.components) || existing.components.length === 0) {
            formsRef.current[el.id] = defaultFormSchema(el.businessObject?.name || initial.name);
          }
        });
      } catch {
        /* алгасах */
      }
    });

    const eventBus = modeler.get<EventBus>('eventBus');
    eventBus.on('selection.changed', (e: SelectionEvent) => {
      const sel = e.newSelection?.[0];
      setSelected(sel ? { id: sel.id, type: sel.type, sourceType: sel.source?.type } : null);
    });

    return () => {
      modeler.destroy();
      modelerRef.current = null;
    };
    // initial-ийг нэг л удаа ачаална (mount).
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Сонгосон user task бүрд form-js editor-ийг холбоно. Сонголт солигдоход
  // өмнөх схемийг хадгалж, шинийг ачаална.
  // Сонголт солигдоход маягтын горимыг (inline / shared) formsRef-ээс тогтооно.
  useEffect(() => {
    const sid = sharedFormId(selectedTaskId ? formsRef.current[selectedTaskId] : undefined);
    setFormMode(sid ? 'shared' : 'inline');
    setSharedSel(sid ?? '');
  }, [selectedTaskId]);

  // Зөвхөн "Энэ процессод зурах" (inline) горимд form-js editor-ийг холбоно.
  // Shared горимд editor үүсгэхгүй (лавлагаа л хадгална).
  useEffect(() => {
    flushFormEditor();
    currentTaskRef.current = selectedTaskId;
    if (selectedTaskId && formMode === 'inline' && formHostRef.current) {
      const existing = formsRef.current[selectedTaskId];
      // Форм байхгүй ЭСВЭЛ shared лавлагаа байсан бол анхдагч inline schema суулгана.
      if (!existing || sharedFormId(existing)) {
        formsRef.current[selectedTaskId] = defaultFormSchema(name);
      }
      const editor = new FormEditor({ container: formHostRef.current });
      formEditorRef.current = editor;
      void editor.importSchema(formsRef.current[selectedTaskId]);
    }
    return () => {
      flushFormEditor();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedTaskId, formMode]);

  // Хуваалцсан формуудыг (form library) modeler-ийн сонголтод нэг удаа татна.
  useEffect(() => {
    let alive = true;
    fetch('/api/bpm/forms', { method: 'GET' })
      .then((r) => r.json())
      .then((b) => { if (alive && b?.ok && Array.isArray(b.data)) setSharedForms(b.data as BpmForm[]); })
      .catch(() => {});
    return () => { alive = false; };
  }, []);

  // Хуваалцсан форм сонгоход тухайн task-д лавлагаа хадгална (inline schema биш).
  const pickSharedForm = (id: string) => {
    if (!selectedTaskId) return;
    setSharedSel(id);
    if (id) formsRef.current[selectedTaskId] = sharedFormRef(id);
  };

  // Идэвхтэй form editor-ийн схемийг formsRef-д хадгалж, editor-ийг устгана.
  function flushFormEditor() {
    if (formEditorRef.current && currentTaskRef.current) {
      formsRef.current[currentTaskRef.current] = formEditorRef.current.saveSchema() as FormSchema;
      formEditorRef.current.destroy();
      formEditorRef.current = null;
    }
  }

  const save = async () => {
    setError('');
    if (!name.trim()) {
      setError(T('bpm.modeler.nameRequired'));
      return;
    }
    const modeler = modelerRef.current;
    if (!modeler) return;

    // Идэвхтэй form editor-ийн өөрчлөлтийг хадгална.
    if (formEditorRef.current && currentTaskRef.current) {
      formsRef.current[currentTaskRef.current] = formEditorRef.current.saveSchema() as FormSchema;
    }

    setBusy(true);
    let xml = '';
    try {
      const out = await modeler.saveXML({ format: true });
      // Маягт, service HTTP, gateway нөхцөлийг стандарт extension element-ээр
      // .bpmn дотор embed хийнэ.
      xml = embedAll(out.xml, formsRef.current, servicesRef.current, conditionsRef.current);
    } catch {
      setBusy(false);
      setError(T('bpm.modeler.saveError'));
      return;
    }

    const payload = {
      name: name.trim(),
      description: initial.description,
      bpmn: xml,
      status: 'draft',
    };
    const res = initial.id
      ? await putJSON(`/api/bpm/processes/${initial.id}`, payload)
      : await postJSON('/api/bpm/processes', payload);
    setBusy(false);
    if (res.ok) {
      router.push('/admin/bpm');
      router.refresh();
      return;
    }
    setError(res.message || T('bpm.modeler.saveError'));
  };

  return (
    <div className={`bpm-modeler${fullscreen ? ' is-fullscreen' : ''}`}>
      <div className="bpm-modeler__toolbar">
        <button className="btn btn--ghost" type="button" onClick={() => router.push('/admin/bpm')}>
          <ArrowLeft size={16} strokeWidth={2} />
          <span>{T('bpm.modeler.back')}</span>
        </button>
        <div className="bpm-modeler__name">
          <input
            className="input"
            placeholder={T('bpm.modeler.namePlaceholder')}
            value={name}
            onChange={(e) => setName(e.target.value)}
            aria-label={T('bpm.modeler.name')}
          />
        </div>
        <button
          className="btn btn--ghost bpm-icon-only"
          type="button"
          aria-label={T(fullscreen ? 'bpm.modeler.exitFullscreen' : 'bpm.modeler.fullscreen')}
          title={T(fullscreen ? 'bpm.modeler.exitFullscreen' : 'bpm.modeler.fullscreen')}
          onClick={() => setFullscreen((f) => !f)}
        >
          {fullscreen ? <Minimize2 size={16} strokeWidth={2} /> : <Maximize2 size={16} strokeWidth={2} />}
        </button>
        <button className="btn btn--primary" type="button" onClick={save} disabled={busy}>
          {busy ? <Loader2 size={16} strokeWidth={2} className="spin" /> : <Save size={16} strokeWidth={2} />}
          <span>{busy ? T('bpm.modeler.saving') : T('bpm.modeler.save')}</span>
        </button>
      </div>

      {error && <div className="alert alert--danger" role="alert">{error}</div>}

      <div className="bpm-modeler__body">
        <div className="bpm-canvas">
          <div ref={canvasRef} className="bpm-canvas__host" />
          <div className="bpm-canvas__zoom">
            <button className="bpm-zoom-btn" type="button" aria-label={T('bpm.modeler.fit')} title={T('bpm.modeler.fit')} onClick={fitView}>
              <Maximize size={15} strokeWidth={2} />
            </button>
            <button className="bpm-zoom-btn" type="button" aria-label={T('bpm.modeler.zoomIn')} title={T('bpm.modeler.zoomIn')} onClick={() => zoomBy(0.2)}>
              <Plus size={15} strokeWidth={2} />
            </button>
            <button className="bpm-zoom-btn" type="button" aria-label={T('bpm.modeler.zoomOut')} title={T('bpm.modeler.zoomOut')} onClick={() => zoomBy(-0.2)}>
              <Minus size={15} strokeWidth={2} />
            </button>
          </div>
        </div>

        <aside className="bpm-form-panel">
          {selected?.type === 'bpmn:UserTask' ? (
            <>
              <div className="bpm-form-panel__head">
                <FileText size={15} strokeWidth={2} />
                <span>{T('bpm.modeler.designForm')}</span>
              </div>
              {/* Маягтын эх сурвалж: энэ процессод зурах эсвэл хуваалцсан сангаас */}
              <div className="bpm-form-source">
                <label className="bpm-form-source__opt">
                  <input
                    type="radio"
                    name="formSource"
                    checked={formMode === 'inline'}
                    onChange={() => setFormMode('inline')}
                  />
                  <span>{T('bpm.modeler.formInline')}</span>
                </label>
                <label className="bpm-form-source__opt">
                  <input
                    type="radio"
                    name="formSource"
                    checked={formMode === 'shared'}
                    onChange={() => { setFormMode('shared'); if (sharedSel) pickSharedForm(sharedSel); }}
                  />
                  <span>{T('bpm.modeler.formShared')}</span>
                </label>
              </div>
              {formMode === 'shared' ? (
                <div className="bpm-form-panel__shared">
                  <select
                    className="input"
                    value={sharedSel}
                    onChange={(e) => pickSharedForm(e.target.value)}
                  >
                    <option value="">{T('bpm.modeler.formPick')}</option>
                    {sharedForms.map((f) => (
                      <option key={f.id} value={f.id}>{f.name}</option>
                    ))}
                  </select>
                  {sharedForms.length === 0 && (
                    <p className="muted bpm-form-panel__sharedhint">{T('bpm.modeler.formNone')}</p>
                  )}
                </div>
              ) : (
                <div ref={formHostRef} className="bpm-form-panel__editor" />
              )}
            </>
          ) : selected?.type === 'bpmn:ServiceTask' ? (
            <ServiceConfigPanel key={selected.id} nodeId={selected.id} servicesRef={servicesRef} />
          ) : isGatewayFlow(selected) ? (
            <ConditionPanel key={selected!.id} flowId={selected!.id} conditionsRef={conditionsRef} />
          ) : (
            <div className="bpm-form-panel__empty muted">
              <MousePointerClick size={20} strokeWidth={1.6} />
              <p>{T('bpm.modeler.selectHint')}</p>
            </div>
          )}
        </aside>
      </div>
    </div>
  );
}

/** serviceTask-ийн HTTP тохиргооны самбар. */
function ServiceConfigPanel({
  nodeId,
  servicesRef,
}: {
  nodeId: string;
  servicesRef: MutableRefObject<Record<string, BpmService>>;
}) {
  const { T } = useT();
  const [cfg, setCfg] = useState<BpmService>(
    () => servicesRef.current[nodeId] ?? { method: 'GET', url: '', body: '', resultVar: '' },
  );
  const update = (patch: Partial<BpmService>) => {
    const next = { ...cfg, ...patch };
    setCfg(next);
    servicesRef.current[nodeId] = next;
  };
  return (
    <div className="bpm-cfg">
      <div className="bpm-form-panel__head">
        <Webhook size={15} strokeWidth={2} />
        <span>{T('bpm.svc.title')}</span>
      </div>
      <div className="bpm-cfg__body">
        <p className="muted" style={{ fontSize: 13 }}>{T('bpm.svc.hint')}</p>
        <div className="field">
          <label className="field__label">{T('bpm.svc.method')}</label>
          <select className="select" value={cfg.method} onChange={(e) => update({ method: e.target.value })}>
            {['GET', 'POST', 'PUT', 'PATCH', 'DELETE'].map((m) => (
              <option key={m} value={m}>{m}</option>
            ))}
          </select>
        </div>
        <div className="field">
          <label className="field__label">{T('bpm.svc.url')}</label>
          <input
            className="input mono"
            placeholder="https://api…/path?x=${var}"
            value={cfg.url}
            onChange={(e) => update({ url: e.target.value })}
          />
        </div>
        <div className="field">
          <label className="field__label">{T('bpm.svc.body')}</label>
          <textarea
            className="input mono"
            rows={4}
            placeholder={'{"amount": ${amount}}'}
            value={cfg.body}
            onChange={(e) => update({ body: e.target.value })}
          />
        </div>
        <div className="field">
          <label className="field__label">{T('bpm.svc.resultVar')}</label>
          <input
            className="input mono"
            placeholder="result"
            value={cfg.resultVar}
            onChange={(e) => update({ resultVar: e.target.value })}
          />
        </div>
      </div>
    </div>
  );
}

/** Gateway-ийн салааны нөхцөлийн самбар. */
function ConditionPanel({
  flowId,
  conditionsRef,
}: {
  flowId: string;
  conditionsRef: MutableRefObject<Record<string, string>>;
}) {
  const { T } = useT();
  const [cond, setCond] = useState<string>(() => conditionsRef.current[flowId] ?? '');
  const update = (v: string) => {
    setCond(v);
    conditionsRef.current[flowId] = v;
  };
  return (
    <div className="bpm-cfg">
      <div className="bpm-form-panel__head">
        <GitBranch size={15} strokeWidth={2} />
        <span>{T('bpm.cond.title')}</span>
      </div>
      <div className="bpm-cfg__body">
        <p className="muted" style={{ fontSize: 13 }}>{T('bpm.cond.hint')}</p>
        <div className="field">
          <label className="field__label">{T('bpm.cond.expr')}</label>
          <input
            className="input mono"
            placeholder="amount >= 1000"
            value={cond}
            onChange={(e) => update(e.target.value)}
          />
        </div>
        <p className="muted" style={{ fontSize: 12 }}>{T('bpm.cond.help')}</p>
      </div>
    </div>
  );
}

/** postJSON-ийн PUT хувилбар (client.ts зөвхөн POST өгдөг). */
async function putJSON(path: string, body: unknown) {
  try {
    const res = await fetch(path, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    let data: { ok?: boolean; status?: number; message?: string } | null = null;
    try {
      data = await res.json();
    } catch {
      /* хоосон body */
    }
    return { ok: data?.ok ?? res.ok, status: data?.status ?? res.status, message: data?.message };
  } catch {
    return { ok: false, status: 0, message: 'Network error' };
  }
}
