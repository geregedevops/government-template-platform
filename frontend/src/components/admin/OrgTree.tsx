"use client";

import { useCallback, useEffect, useState } from 'react';
import { Building2, Plus, Pencil, Trash2, Loader2, Check, X } from 'lucide-react';
import { useT } from '@/lib/useT';
import type { DictKey } from '@/lib/i18n';

interface Org {
  id: string;
  parent_id: string;
  path: string;
  name: string;
  kind: 'root' | 'ministry' | 'agency' | 'soe';
}
interface Node extends Org {
  children: Node[];
  depth: number;
}

const KIND_LABEL: Record<Org['kind'], DictKey> = {
  root: 'org.kind.root',
  ministry: 'org.kind.ministry',
  agency: 'org.kind.agency',
  soe: 'org.kind.soe',
};

const ROOT_ID = '00000000-0000-0000-0000-000000000001';

export default function OrgTree() {
  const { T } = useT();
  const [items, setItems] = useState<Org[] | null>(null);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  // Идэвхтэй маягт: тухайн node-д хүүхэд нэмэх эсвэл засах.
  const [form, setForm] = useState<{ mode: 'add' | 'edit'; org: Org } | null>(null);
  const [name, setName] = useState('');
  const [kind, setKind] = useState<Org['kind']>('agency');

  const load = useCallback(async () => {
    setError('');
    try {
      const res = await fetch('/api/orgs', { method: 'GET' });
      const body = (await res.json()) as { ok: boolean; data?: Org[]; message?: string };
      if (body.ok && Array.isArray(body.data)) setItems(body.data);
      else { setItems([]); setError(body.message || T('org.loadError')); }
    } catch {
      setItems([]); setError(T('org.loadError'));
    }
  }, [T]);

  useEffect(() => { void load(); }, [load]);

  // Path-аар эрэмбэлэгдсэн жагсаалтаас мод барина.
  const roots: Node[] = (() => {
    if (!items) return [];
    const byId = new Map<string, Node>();
    for (const o of items) byId.set(o.id, { ...o, children: [], depth: 0 });
    const tops: Node[] = [];
    for (const n of byId.values()) {
      const parent = n.parent_id ? byId.get(n.parent_id) : undefined;
      if (parent) parent.children.push(n);
      else tops.push(n);
    }
    return tops;
  })();

  const openAdd = (org: Org) => { setForm({ mode: 'add', org }); setName(''); setKind('agency'); setError(''); };
  const openEdit = (org: Org) => { setForm({ mode: 'edit', org }); setName(org.name); setKind(org.kind); setError(''); };

  const submit = async () => {
    if (!form) return;
    if (!name.trim()) { setError(T('org.nameRequired')); return; }
    setBusy(true);
    setError('');
    try {
      const isAdd = form.mode === 'add';
      const res = await fetch(isAdd ? '/api/orgs' : `/api/orgs/${form.org.id}`, {
        method: isAdd ? 'POST' : 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(isAdd
          ? { parent_id: form.org.id, name: name.trim(), kind }
          : { name: name.trim(), kind }),
      });
      const body = (await res.json()) as { ok: boolean; message?: string };
      if (!res.ok || !body.ok) { setError(body.message || T('org.saveError')); return; }
      setForm(null);
      await load();
    } catch {
      setError(T('org.saveError'));
    } finally {
      setBusy(false);
    }
  };

  const remove = async (org: Org) => {
    if (!window.confirm(T('org.deleteConfirm'))) return;
    setBusy(true);
    setError('');
    try {
      const res = await fetch(`/api/orgs/${org.id}`, { method: 'DELETE' });
      const body = (await res.json().catch(() => ({ ok: res.ok }))) as { ok?: boolean; message?: string };
      if (!res.ok || !body.ok) { setError(body.message || T('org.deleteError')); return; }
      await load();
    } catch {
      setError(T('org.deleteError'));
    } finally {
      setBusy(false);
    }
  };

  const flat: Node[] = [];
  const walk = (nodes: Node[], depth: number) => {
    for (const n of nodes) { n.depth = depth; flat.push(n); walk(n.children, depth + 1); }
  };
  walk(roots, 0);

  return (
    <div className="orgtree">
      {error && <div className="alert alert--danger" role="alert">{error}</div>}

      {items === null ? (
        <div className="bpm-run__state muted"><Loader2 size={18} strokeWidth={2} className="spin" /><span>{T('org.loading')}</span></div>
      ) : (
        <div className="card orgtree__list">
          {flat.map((n) => (
            <div key={n.id}>
              <div className="orgtree__row" style={{ paddingLeft: 12 + n.depth * 22 }}>
                <Building2 size={15} strokeWidth={2} className="orgtree__icon" />
                <span className="orgtree__name">{n.name}</span>
                <span className="chip chip--neutral">{T(KIND_LABEL[n.kind])}</span>
                <div className="orgtree__actions">
                  <button className="btn btn--ghost btn--sm" type="button" onClick={() => openAdd(n)} title={T('org.addChild')}>
                    <Plus size={14} strokeWidth={2} />
                  </button>
                  {n.id !== ROOT_ID && (
                    <>
                      <button className="btn btn--ghost btn--sm" type="button" onClick={() => openEdit(n)} title={T('org.edit')}>
                        <Pencil size={14} strokeWidth={2} />
                      </button>
                      <button className="btn btn--ghost btn--sm" type="button" onClick={() => remove(n)} disabled={busy} title={T('org.delete')}>
                        <Trash2 size={14} strokeWidth={2} />
                      </button>
                    </>
                  )}
                </div>
              </div>
              {form && form.org.id === n.id && (
                <div className="orgtree__form" style={{ paddingLeft: 12 + (n.depth + (form.mode === 'add' ? 1 : 0)) * 22 }}>
                  <span className="muted orgtree__formlabel">
                    {form.mode === 'add' ? T('org.addUnder') + ' ' + n.name : T('org.editing')}
                  </span>
                  <input
                    className="input"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder={T('org.namePlaceholder')}
                    maxLength={200}
                    autoFocus
                  />
                  <select className="input" value={kind} onChange={(e) => setKind(e.target.value as Org['kind'])}>
                    {form.mode === 'edit' && <option value="root">{T('org.kind.root')}</option>}
                    <option value="ministry">{T('org.kind.ministry')}</option>
                    <option value="agency">{T('org.kind.agency')}</option>
                    <option value="soe">{T('org.kind.soe')}</option>
                  </select>
                  <button className="btn btn--primary btn--sm" type="button" onClick={submit} disabled={busy}>
                    {busy ? <Loader2 size={14} strokeWidth={2} className="spin" /> : <Check size={14} strokeWidth={2} />}
                  </button>
                  <button className="btn btn--secondary btn--sm" type="button" onClick={() => setForm(null)} disabled={busy}>
                    <X size={14} strokeWidth={2} />
                  </button>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
