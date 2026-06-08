"use client";

import { useCallback, useEffect, useRef, useState } from 'react';
import '@bpmn-io/form-js/dist/assets/form-js-editor.css';
import '@bpmn-io/form-js/dist/assets/properties-panel.css';
import { Plus, Pencil, Trash2, Loader2, Save, X, FileText } from 'lucide-react';
import { useT } from '@/lib/useT';
import { emptyFormSchema, type BpmForm, type FormSchema } from '@/lib/bpm';

interface FormEditorInstance {
  importSchema: (schema: unknown) => Promise<unknown>;
  saveSchema: () => Record<string, unknown>;
  destroy: () => void;
}

type Editing = { id: string | null; name: string; schema: FormSchema } | null;

/**
 * Хуваалцсан форм сан — олон процесс дунд ашиглах формуудыг form-js editor-оор
 * үүсгэх/засах/устгах. form-js нь DOM шаарддаг тул editor-ийг ЗӨВХӨН клиент
 * талд динамикаар ачаална.
 */
export default function FormsManager() {
  const { T } = useT();
  const [items, setItems] = useState<BpmForm[] | null>(null);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const [deleting, setDeleting] = useState<string | null>(null);
  const [editing, setEditing] = useState<Editing>(null);

  const hostRef = useRef<HTMLDivElement>(null);
  const editorRef = useRef<FormEditorInstance | null>(null);

  const load = useCallback(async () => {
    setError('');
    try {
      const res = await fetch('/api/bpm/forms', { method: 'GET' });
      const body = (await res.json()) as { ok: boolean; data?: BpmForm[]; message?: string };
      if (body.ok && Array.isArray(body.data)) setItems(body.data);
      else { setItems([]); setError(body.message || T('bpm.forms.loadError')); }
    } catch {
      setItems([]); setError(T('bpm.forms.loadError'));
    }
  }, [T]);

  useEffect(() => { void load(); }, [load]);

  // form-js editor-ийг засварлах горимд динамикаар холбоно.
  useEffect(() => {
    if (!editing) return;
    let destroyed = false;
    let editor: FormEditorInstance | null = null;
    (async () => {
      if (!hostRef.current) return;
      try {
        const mod = await import('@bpmn-io/form-js');
        if (destroyed || !hostRef.current) return;
        editor = new mod.FormEditor({ container: hostRef.current }) as unknown as FormEditorInstance;
        editorRef.current = editor;
        const safe = editing.schema && typeof editing.schema === 'object'
          ? editing.schema
          : emptyFormSchema();
        await editor.importSchema(safe);
      } catch {
        if (!destroyed) setError(T('bpm.forms.editorError'));
      }
    })();
    return () => {
      destroyed = true;
      if (editor) { try { editor.destroy(); } catch { /* алгасах */ } }
      editorRef.current = null;
    };
  }, [editing, T]);

  const save = async () => {
    if (!editing) return;
    if (!editing.name.trim()) { setError(T('bpm.forms.nameRequired')); return; }
    const schema = editorRef.current?.saveSchema() ?? editing.schema;
    setBusy(true);
    setError('');
    try {
      const isNew = !editing.id;
      const res = await fetch(isNew ? '/api/bpm/forms' : `/api/bpm/forms/${editing.id}`, {
        method: isNew ? 'POST' : 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: editing.name.trim(), schema }),
      });
      const body = (await res.json()) as { ok: boolean; message?: string };
      if (!res.ok || !body.ok) { setError(body.message || T('bpm.forms.saveError')); return; }
      setEditing(null);
      await load();
    } catch {
      setError(T('bpm.forms.saveError'));
    } finally {
      setBusy(false);
    }
  };

  const remove = async (id: string) => {
    if (!window.confirm(T('bpm.forms.deleteConfirm'))) return;
    setDeleting(id);
    setError('');
    try {
      const res = await fetch(`/api/bpm/forms/${id}`, { method: 'DELETE' });
      const body = (await res.json().catch(() => ({ ok: res.ok }))) as { ok?: boolean; message?: string };
      if (!res.ok || !body.ok) { setError(body.message || T('bpm.forms.deleteError')); return; }
      setItems((prev) => (prev ? prev.filter((f) => f.id !== id) : prev));
    } catch {
      setError(T('bpm.forms.deleteError'));
    } finally {
      setDeleting(null);
    }
  };

  if (editing) {
    return (
      <div className="bpm-forms">
        {error && <div className="alert alert--danger" role="alert">{error}</div>}
        <div className="bpm-forms__editbar">
          <input
            className="input"
            value={editing.name}
            onChange={(e) => setEditing({ ...editing, name: e.target.value })}
            placeholder={T('bpm.forms.namePlaceholder')}
            maxLength={200}
          />
          <button className="btn btn--secondary" type="button" onClick={() => setEditing(null)} disabled={busy}>
            <X size={16} strokeWidth={2} /><span>{T('bpm.forms.cancel')}</span>
          </button>
          <button className="btn btn--primary" type="button" onClick={save} disabled={busy}>
            {busy ? <Loader2 size={16} strokeWidth={2} className="spin" /> : <Save size={16} strokeWidth={2} />}
            <span>{T('bpm.forms.save')}</span>
          </button>
        </div>
        <div ref={hostRef} className="bpm-form-panel__editor bpm-forms__host" />
      </div>
    );
  }

  return (
    <div className="bpm-forms">
      {error && <div className="alert alert--danger" role="alert">{error}</div>}
      <div className="bpm-list__head">
        <button
          className="btn btn--primary"
          type="button"
          onClick={() => setEditing({ id: null, name: '', schema: emptyFormSchema() })}
        >
          <Plus size={16} strokeWidth={2} /><span>{T('bpm.forms.add')}</span>
        </button>
      </div>

      {items === null ? (
        <div className="bpm-run__state muted"><Loader2 size={18} strokeWidth={2} className="spin" /><span>{T('bpm.forms.loading')}</span></div>
      ) : items.length === 0 ? (
        <div className="card bpm-empty"><FileText size={20} strokeWidth={1.6} /><p>{T('bpm.forms.empty')}</p></div>
      ) : (
        <div className="bpm-grid">
          {items.map((f) => (
            <article key={f.id} className="card bpm-card">
              <div className="bpm-card__body">
                <h3 className="bpm-card__title">{f.name}</h3>
                <span className="chip chip--neutral mono">{(f.schema?.components?.length ?? 0)} талбар</span>
              </div>
              <div className="bpm-card__actions">
                <button className="btn btn--secondary btn--sm" type="button" onClick={() => setEditing({ id: f.id, name: f.name, schema: f.schema ?? emptyFormSchema() })}>
                  <Pencil size={14} strokeWidth={2} /><span>{T('bpm.forms.edit')}</span>
                </button>
                <button className="btn btn--ghost btn--sm" type="button" onClick={() => remove(f.id)} disabled={deleting === f.id}>
                  {deleting === f.id ? <Loader2 size={14} strokeWidth={2} className="spin" /> : <Trash2 size={14} strokeWidth={2} />}
                  <span>{T('bpm.forms.delete')}</span>
                </button>
              </div>
            </article>
          ))}
        </div>
      )}
    </div>
  );
}
