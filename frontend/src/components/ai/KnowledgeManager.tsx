"use client";

import { useCallback, useEffect, useState } from 'react';
import { Plus, Pencil, Trash2, Loader2, BookOpen, X, Save } from 'lucide-react';
import { useT } from '@/lib/useT';

interface KnowledgeEntry {
  id: string;
  title: string;
  content: string;
  owner_email?: string;
  created_at: string;
  updated_at: string | null;
}
interface ApiResponse<T> {
  ok: boolean;
  data?: T;
  message?: string;
}

// editing: null = форм хаалттай, 'new' = шинэ, бусад = тухайн id-г засаж байна.
type Editing = string | 'new' | null;

export default function KnowledgeManager() {
  const { T } = useT();
  const [items, setItems] = useState<KnowledgeEntry[] | null>(null);
  const [editing, setEditing] = useState<Editing>(null);
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  const load = useCallback(async () => {
    setError('');
    try {
      const res = await fetch('/api/ai/knowledge/all', { method: 'GET' });
      const body = (await res.json()) as ApiResponse<KnowledgeEntry[]>;
      if (body.ok) setItems(body.data ?? []);
      else {
        setItems([]);
        setError(body.message || T('kb.loadError'));
      }
    } catch {
      setItems([]);
      setError(T('kb.loadError'));
    }
  }, [T]);

  useEffect(() => {
    void load();
  }, [load]);

  const openNew = () => {
    setEditing('new');
    setTitle('');
    setContent('');
    setError('');
  };
  const openEdit = (k: KnowledgeEntry) => {
    setEditing(k.id);
    setTitle(k.title);
    setContent(k.content);
    setError('');
  };
  const cancel = () => {
    setEditing(null);
    setTitle('');
    setContent('');
  };

  const submit = async () => {
    if (!content.trim()) {
      setError(T('kb.contentRequired'));
      return;
    }
    setBusy(true);
    setError('');
    const payload = { title: title.trim(), content: content.trim() };
    const url = editing === 'new' ? '/api/ai/knowledge' : `/api/ai/knowledge/${editing}`;
    const method = editing === 'new' ? 'POST' : 'PUT';
    try {
      const res = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      const body = (await res.json()) as ApiResponse<KnowledgeEntry>;
      setBusy(false);
      if (body.ok) {
        cancel();
        void load();
        return;
      }
      setError(body.message || T('kb.saveError'));
    } catch {
      setBusy(false);
      setError(T('kb.saveError'));
    }
  };

  const remove = async (id: string) => {
    if (!window.confirm(T('kb.deleteConfirm'))) return;
    try {
      const res = await fetch(`/api/ai/knowledge/${id}`, { method: 'DELETE' });
      const body = (await res.json()) as ApiResponse<unknown>;
      if (body.ok) void load();
      else setError(body.message || T('kb.deleteError'));
    } catch {
      setError(T('kb.deleteError'));
    }
  };

  return (
    <div className="kb">
      <div className="kb__head">
        {editing === null && (
          <button className="btn btn--primary" type="button" onClick={openNew}>
            <Plus size={16} strokeWidth={2} />
            <span>{T('kb.add')}</span>
          </button>
        )}
      </div>

      {error && <div className="alert alert--danger" role="alert">{error}</div>}

      {editing !== null && (
        <div className="card kb-form">
          <div className="kb-form__head">
            <BookOpen size={15} strokeWidth={2} />
            <span>{editing === 'new' ? T('kb.add') : T('kb.edit')}</span>
          </div>
          <div className="field">
            <label className="field__label" htmlFor="kb-title">{T('kb.titleLabel')}</label>
            <input
              id="kb-title"
              className="input"
              value={title}
              maxLength={200}
              placeholder={T('kb.titlePlaceholder')}
              onChange={(e) => setTitle(e.target.value)}
            />
          </div>
          <div className="field">
            <label className="field__label" htmlFor="kb-content">{T('kb.contentLabel')}</label>
            <textarea
              id="kb-content"
              className="input kb-form__textarea"
              value={content}
              maxLength={8000}
              rows={6}
              placeholder={T('kb.contentPlaceholder')}
              onChange={(e) => setContent(e.target.value)}
            />
          </div>
          <div className="kb-form__actions">
            <button className="btn btn--primary" type="button" onClick={submit} disabled={busy}>
              {busy ? <Loader2 size={16} strokeWidth={2} className="spin" /> : <Save size={16} strokeWidth={2} />}
              <span>{T('kb.save')}</span>
            </button>
            <button className="btn btn--secondary" type="button" onClick={cancel} disabled={busy}>
              <X size={16} strokeWidth={2} />
              <span>{T('kb.cancel')}</span>
            </button>
          </div>
        </div>
      )}

      {items === null && (
        <div className="bpm-list__loading muted">
          <Loader2 size={16} strokeWidth={2} className="spin" />
          <span>{T('kb.loading')}</span>
        </div>
      )}

      {items !== null && items.length === 0 && editing === null && (
        <div className="card bpm-empty">
          <BookOpen size={22} strokeWidth={1.6} />
          <p className="muted">{T('kb.empty')}</p>
        </div>
      )}

      {items !== null && items.length > 0 && (
        <div className="kb__list">
          {items.map((k) => (
            <article key={k.id} className="card kb-card">
              <div className="kb-card__body">
                {k.title && <h3 className="kb-card__title">{k.title}</h3>}
                {k.owner_email && <span className="chip chip--neutral mono kb-card__owner">{k.owner_email}</span>}
                <p className="kb-card__content">{k.content}</p>
              </div>
              <div className="kb-card__actions">
                <button className="btn btn--secondary btn--sm" type="button" onClick={() => openEdit(k)}>
                  <Pencil size={14} strokeWidth={2} />
                  <span>{T('kb.edit')}</span>
                </button>
                <button className="btn btn--ghost btn--sm bpm-card__danger" type="button" onClick={() => remove(k.id)}>
                  <Trash2 size={14} strokeWidth={2} />
                  <span>{T('kb.delete')}</span>
                </button>
              </div>
            </article>
          ))}
        </div>
      )}
    </div>
  );
}
