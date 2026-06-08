"use client";

import { useCallback, useEffect, useRef, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Plus, Play, Pencil, Trash2, Workflow, Loader2, Upload, FileUp, X, Sparkles, Activity, FileText } from 'lucide-react';
import { useT } from '@/lib/useT';
import { postJSON } from '@/lib/client';
import type { DictKey } from '@/lib/i18n';
import type { BpmProcess } from '@/lib/bpm';

interface ListResponse {
  ok: boolean;
  data?: BpmProcess[];
  message?: string;
}

export default function ProcessList() {
  const { T, lang } = useT();
  const [items, setItems] = useState<BpmProcess[] | null>(null);
  const [error, setError] = useState('');
  const [deleting, setDeleting] = useState<string | null>(null);

  const load = useCallback(async () => {
    setError('');
    try {
      const res = await fetch('/api/bpm/processes', { method: 'GET' });
      const body = (await res.json()) as ListResponse;
      if (body.ok) {
        setItems(body.data ?? []);
      } else {
        setItems([]);
        setError(body.message || T('bpm.loadError'));
      }
    } catch {
      setItems([]);
      setError(T('bpm.loadError'));
    }
  }, [T]);

  useEffect(() => {
    void load();
  }, [load]);

  const remove = async (id: string) => {
    if (!window.confirm(T('bpm.confirmDelete'))) return;
    setDeleting(id);
    setError('');
    try {
      // Хариуг шалгана — backend амжилтгүй болоход мөрийг UI-аас алга
      // болгохгүй (өмнө нь чимээгүй алга болоод дараагийн ачаалалт дээр
      // эргэж гарч ирдэг байсан).
      const res = await fetch(`/api/bpm/processes/${id}`, { method: 'DELETE' });
      const body = (await res.json().catch(() => ({ ok: res.ok }))) as { ok?: boolean; message?: string };
      if (!res.ok || !body.ok) {
        setError(body.message || T('bpm.deleteError'));
        return;
      }
      setItems((prev) => (prev ? prev.filter((p) => p.id !== id) : prev));
    } catch {
      setError(T('bpm.deleteError'));
    } finally {
      setDeleting(null);
    }
  };

  return (
    <div className="bpm-list">
      <div className="bpm-list__head">
        <Link className="btn btn--secondary" href="/admin/bpm/forms">
          <FileText size={16} strokeWidth={2} />
          <span>{T('bpm.forms.link')}</span>
        </Link>
        <AiGenerate />
        <ImportBpmn onImported={load} />
        <Link className="btn btn--primary" href="/admin/bpm/modeler/new">
          <Plus size={16} strokeWidth={2} />
          <span>{T('bpm.new')}</span>
        </Link>
      </div>

      {error && <div className="alert alert--danger" role="alert">{error}</div>}

      {items === null && (
        <div className="bpm-list__loading muted">
          <Loader2 size={16} strokeWidth={2} className="spin" />
          <span>{T('bpm.run.loading')}</span>
        </div>
      )}

      {items !== null && items.length === 0 && !error && (
        <div className="card bpm-empty">
          <Workflow size={22} strokeWidth={1.6} />
          <p className="muted">{T('bpm.empty')}</p>
        </div>
      )}

      {items !== null && items.length > 0 && (
        <div className="bpm-grid">
          {items.map((p) => (
            <article key={p.id} className="card bpm-card">
              <div className="bpm-card__head">
                <h3 className="bpm-card__title" title={p.name}>{p.name}</h3>
                <span className="chip">{T(statusKey(p.status))}</span>
              </div>
              {p.description && <p className="muted bpm-card__desc">{p.description}</p>}
              <div className="defrow">
                <span className="defrow__label">{T('bpm.created')}</span>
                <span className="defrow__value mono">{formatDate(p.created_at, lang)}</span>
              </div>
              <div className="bpm-card__actions">
                <Link className="btn btn--primary btn--sm" href={`/admin/bpm/run/${p.id}`}>
                  <Play size={14} strokeWidth={2} />
                  <span>{T('bpm.run')}</span>
                </Link>
                <Link className="btn btn--secondary btn--sm" href={`/admin/bpm/modeler/${p.id}`}>
                  <Pencil size={14} strokeWidth={2} />
                  <span>{T('bpm.edit')}</span>
                </Link>
                <Link className="btn btn--ghost btn--sm" href={`/admin/bpm/runs/${p.id}`}>
                  <Activity size={14} strokeWidth={2} />
                  <span>{T('bpm.runs')}</span>
                </Link>
                <button
                  className="btn btn--ghost btn--sm"
                  type="button"
                  onClick={() => remove(p.id)}
                  disabled={deleting === p.id}
                >
                  <Trash2 size={14} strokeWidth={2} />
                  <span>{T('bpm.delete')}</span>
                </button>
              </div>
            </article>
          ))}
        </div>
      )}
    </div>
  );
}

/**
 * .bpmn файл оруулах хяналт: товч дарахад inline самбар нээгдэж, файл сонгоод
 * нэр өгч хадгална. Файлын агуулгыг client талд уншиж, одоо байгаа
 * POST /api/bpm/processes руу { bpmn, forms:{} } болгон илгээнэ (backend нь
 * BPMN-ийг шалгана). Хадгалсны дараа жагсаалтыг дахин ачаална.
 */
function ImportBpmn({ onImported }: { onImported: () => void | Promise<void> }) {
  const { T } = useT();
  const [open, setOpen] = useState(false);
  const [name, setName] = useState('');
  const [bpmn, setBpmn] = useState('');
  const [fileName, setFileName] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);

  const reset = () => {
    setName('');
    setBpmn('');
    setFileName('');
    setError('');
    if (inputRef.current) inputRef.current.value = '';
  };

  const close = () => {
    setOpen(false);
    reset();
  };

  const onFile = async (e: React.ChangeEvent<HTMLInputElement>) => {
    setError('');
    const file = e.target.files?.[0];
    if (!file) return;
    const text = await file.text();
    // Хөнгөн client-side шалгалт — backend нь бүрэн шалгана.
    if (!/<\s*(bpmn:)?definitions/i.test(text)) {
      setError(T('bpm.import.invalidFile'));
      setBpmn('');
      return;
    }
    setBpmn(text);
    setFileName(file.name);
    if (!name.trim()) setName(file.name.replace(/\.(bpmn|xml)$/i, ''));
  };

  const save = async () => {
    setError('');
    if (!bpmn) {
      setError(T('bpm.import.needFile'));
      return;
    }
    if (!name.trim()) {
      setError(T('bpm.modeler.nameRequired'));
      return;
    }
    setBusy(true);
    // Оруулсан файл нь өөрөө цэвэр .bpmn (маягтууд нь дотроо embed хийгдсэн
    // байж болно) — шууд хадгална. Backend нь BPMN-ийг шалгана.
    const res = await postJSON('/api/bpm/processes', {
      name: name.trim(),
      description: '',
      bpmn,
      status: 'draft',
    });
    setBusy(false);
    if (res.ok) {
      close();
      await onImported();
      return;
    }
    setError(res.message || T('bpm.import.invalidFile'));
  };

  if (!open) {
    return (
      <button className="btn btn--secondary" type="button" onClick={() => setOpen(true)}>
        <Upload size={16} strokeWidth={2} />
        <span>{T('bpm.import')}</span>
      </button>
    );
  }

  return (
    <div className="card bpm-import">
      <div className="bpm-import__head">
        <span className="field__label">{T('bpm.import.title')}</span>
        <button className="bpm-icon-btn" type="button" aria-label={T('bpm.import.cancel')} onClick={close}>
          <X size={14} strokeWidth={2} />
        </button>
      </div>

      <div className="field">
        <label className="field__label" htmlFor="bpmn-file">{T('bpm.import.pickFile')}</label>
        <input
          ref={inputRef}
          id="bpmn-file"
          type="file"
          accept=".bpmn,.xml,application/xml,text/xml"
          className="input"
          onChange={onFile}
        />
        {fileName && (
          <span className="muted" style={{ fontSize: 13, display: 'inline-flex', alignItems: 'center', gap: 6 }}>
            <FileUp size={13} strokeWidth={2} />{fileName}
          </span>
        )}
      </div>

      <div className="field">
        <label className="field__label" htmlFor="bpmn-name">{T('bpm.import.name')}</label>
        <input
          id="bpmn-name"
          className="input"
          placeholder={T('bpm.modeler.namePlaceholder')}
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
      </div>

      {error && <div className="alert alert--danger" role="alert">{error}</div>}

      <div className="bpm-import__actions">
        <button className="btn btn--primary" type="button" onClick={save} disabled={busy}>
          {busy ? <Loader2 size={16} strokeWidth={2} className="spin" /> : <Upload size={16} strokeWidth={2} />}
          <span>{busy ? T('bpm.modeler.saving') : T('bpm.import.save')}</span>
        </button>
        <button className="btn btn--ghost" type="button" onClick={close} disabled={busy}>
          <span>{T('bpm.import.cancel')}</span>
        </button>
      </div>
    </div>
  );
}

/**
 * AI-аар процесс үүсгэх хяналт: тайлбар бичээд Generate дарахад backend Claude-
 * аар BPMN процесс үүсгэж хадгална; дараа нь modeler руу шилжиж засна.
 */
function AiGenerate() {
  const { T } = useT();
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [desc, setDesc] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  const generate = async () => {
    setError('');
    if (desc.trim().length < 4) {
      setError(T('bpm.ai.needDesc'));
      return;
    }
    setBusy(true);
    try {
      const res = await fetch('/api/bpm/generate', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ description: desc.trim() }),
      });
      const body = (await res.json()) as { ok: boolean; data?: { id?: string }; message?: string };
      if (body.ok && body.data?.id) {
        router.push(`/admin/bpm/modeler/${body.data.id}`);
        return;
      }
      setError(body.message || T('bpm.ai.error'));
    } catch {
      setError(T('bpm.ai.error'));
    } finally {
      setBusy(false);
    }
  };

  if (!open) {
    return (
      <button className="btn btn--secondary" type="button" onClick={() => setOpen(true)}>
        <Sparkles size={16} strokeWidth={2} />
        <span>{T('bpm.ai.button')}</span>
      </button>
    );
  }

  return (
    <div className="card bpm-import">
      <div className="bpm-import__head">
        <span className="field__label">{T('bpm.ai.title')}</span>
        <button className="bpm-icon-btn" type="button" aria-label={T('bpm.import.cancel')} onClick={() => setOpen(false)}>
          <X size={14} strokeWidth={2} />
        </button>
      </div>
      <div className="field">
        <label className="field__label" htmlFor="ai-desc">{T('bpm.ai.label')}</label>
        <textarea
          id="ai-desc"
          className="input"
          rows={3}
          placeholder={T('bpm.ai.placeholder')}
          value={desc}
          onChange={(e) => setDesc(e.target.value)}
          disabled={busy}
        />
      </div>
      {error && <div className="alert alert--danger" role="alert">{error}</div>}
      <div className="bpm-import__actions">
        <button className="btn btn--primary" type="button" onClick={generate} disabled={busy}>
          {busy ? <Loader2 size={16} strokeWidth={2} className="spin" /> : <Sparkles size={16} strokeWidth={2} />}
          <span>{busy ? T('bpm.ai.generating') : T('bpm.ai.generate')}</span>
        </button>
        <button className="btn btn--ghost" type="button" onClick={() => setOpen(false)} disabled={busy}>
          <span>{T('bpm.import.cancel')}</span>
        </button>
      </div>
    </div>
  );
}

function statusKey(status: string): DictKey {
  return status === 'published' ? 'bpm.status.published' : 'bpm.status.draft';
}

function formatDate(iso: string, lang: string): string {
  try {
    return new Date(iso).toLocaleDateString(lang === 'en' ? 'en-US' : 'mn-MN', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  } catch {
    return iso.slice(0, 10);
  }
}
