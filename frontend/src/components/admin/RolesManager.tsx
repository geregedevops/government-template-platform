"use client";

import { useCallback, useEffect, useState } from 'react';
import { Plus, Trash2, Loader2, ShieldHalf, X, Save } from 'lucide-react';
import { useT } from '@/lib/useT';

interface Role {
  id: number;
  key: string;
  name: string;
  description: string;
  is_system: boolean;
  permissions: string[];
}
interface Permission {
  key: string;
  label: string;
  category: string;
}
interface ApiResponse<T> {
  ok: boolean;
  data?: T;
  message?: string;
}

export default function RolesManager() {
  const { T } = useT();
  const [roles, setRoles] = useState<Role[] | null>(null);
  const [perms, setPerms] = useState<Permission[]>([]);
  const [error, setError] = useState('');
  const [adding, setAdding] = useState(false);
  const [busy, setBusy] = useState(false);
  const [form, setForm] = useState({ name: '', description: '' });
  // role.id -> локал сонгосон permission set (хадгалахаас өмнө)
  const [draft, setDraft] = useState<Record<number, Set<string>>>({});

  const load = useCallback(async () => {
    setError('');
    try {
      const [rRes, pRes] = await Promise.all([
        fetch('/api/rbac/roles', { method: 'GET' }),
        fetch('/api/rbac/permissions', { method: 'GET' }),
      ]);
      const rBody = (await rRes.json()) as ApiResponse<Role[]>;
      const pBody = (await pRes.json()) as ApiResponse<Permission[]>;
      if (rBody.ok) {
        const list = rBody.data ?? [];
        setRoles(list);
        const d: Record<number, Set<string>> = {};
        for (const role of list) d[role.id] = new Set(role.permissions);
        setDraft(d);
      } else {
        setRoles([]);
        setError(rBody.message || T('roles.loadError'));
      }
      if (pBody.ok) setPerms(pBody.data ?? []);
    } catch {
      setRoles([]);
      setError(T('roles.loadError'));
    }
  }, [T]);

  useEffect(() => {
    void load();
  }, [load]);

  const categories = Array.from(new Set(perms.map((p) => p.category)));

  const toggle = (roleId: number, key: string) => {
    setDraft((prev) => {
      const next = new Set(prev[roleId] ?? []);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return { ...prev, [roleId]: next };
    });
  };

  const savePerms = async (role: Role) => {
    setBusy(true);
    setError('');
    try {
      const res = await fetch(`/api/rbac/roles/${role.id}/permissions`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ permissions: Array.from(draft[role.id] ?? []) }),
      });
      const body = (await res.json()) as ApiResponse<unknown>;
      setBusy(false);
      if (body.ok) void load();
      else setError(body.message || T('roles.saveError'));
    } catch {
      setBusy(false);
      setError(T('roles.saveError'));
    }
  };

  const create = async () => {
    if (!form.name.trim()) {
      setError(T('roles.nameRequired'));
      return;
    }
    setBusy(true);
    setError('');
    try {
      const res = await fetch('/api/rbac/roles', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: form.name, description: form.description, permissions: [] }),
      });
      const body = (await res.json()) as ApiResponse<Role>;
      setBusy(false);
      if (body.ok) {
        setAdding(false);
        setForm({ name: '', description: '' });
        void load();
      } else setError(body.message || T('roles.saveError'));
    } catch {
      setBusy(false);
      setError(T('roles.saveError'));
    }
  };

  const remove = async (role: Role) => {
    if (!window.confirm(T('roles.deleteConfirm'))) return;
    setError('');
    try {
      const res = await fetch(`/api/rbac/roles/${role.id}`, { method: 'DELETE' });
      const body = (await res.json()) as ApiResponse<unknown>;
      if (body.ok) void load();
      else setError(body.message || T('roles.deleteError'));
    } catch {
      setError(T('roles.deleteError'));
    }
  };

  const dirty = (role: Role) => {
    const d = draft[role.id] ?? new Set<string>();
    const orig = new Set(role.permissions);
    if (d.size !== orig.size) return true;
    for (const k of d) if (!orig.has(k)) return true;
    return false;
  };

  return (
    <div className="roles">
      <div className="roles__head">
        {!adding && (
          <button className="btn btn--primary" type="button" onClick={() => { setAdding(true); setError(''); }}>
            <Plus size={16} strokeWidth={2} />
            <span>{T('roles.add')}</span>
          </button>
        )}
      </div>

      {error && <div className="alert alert--danger" role="alert">{error}</div>}

      {adding && (
        <div className="card roles-form">
          <div className="roles-form__head"><ShieldHalf size={15} strokeWidth={2} /><span>{T('roles.add')}</span></div>
          <div className="users-form__grid">
            <div className="field">
              <label className="field__label" htmlFor="r-name">{T('roles.name')}</label>
              <input id="r-name" className="input" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
            </div>
            <div className="field">
              <label className="field__label" htmlFor="r-desc">{T('roles.description')}</label>
              <input id="r-desc" className="input" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
            </div>
          </div>
          <div className="users-form__actions">
            <button className="btn btn--primary" type="button" onClick={create} disabled={busy}>
              {busy ? <Loader2 size={16} strokeWidth={2} className="spin" /> : <Save size={16} strokeWidth={2} />}
              <span>{T('roles.create')}</span>
            </button>
            <button className="btn btn--secondary" type="button" onClick={() => setAdding(false)} disabled={busy}>
              <X size={16} strokeWidth={2} /><span>{T('roles.cancel')}</span>
            </button>
          </div>
        </div>
      )}

      {roles === null && (
        <div className="bpm-list__loading muted"><Loader2 size={16} strokeWidth={2} className="spin" /><span>{T('roles.loading')}</span></div>
      )}

      {roles?.map((role) => {
        const isAdmin = role.key === 'admin';
        const sel = draft[role.id] ?? new Set<string>();
        return (
          <article key={role.id} className="card role-card">
            <div className="role-card__head">
              <div>
                <span className="role-card__name">{role.name}</span>
                <span className="chip chip--neutral mono" style={{ marginLeft: 8 }}>{role.key}</span>
                {role.is_system && <span className="chip" style={{ marginLeft: 6 }}>{T('roles.system')}</span>}
              </div>
              {!role.is_system && (
                <button className="btn btn--ghost btn--sm" type="button" onClick={() => remove(role)}>
                  <Trash2 size={14} strokeWidth={2} /><span>{T('roles.delete')}</span>
                </button>
              )}
            </div>
            {role.description && <p className="muted role-card__desc">{role.description}</p>}

            <div className="role-card__perms">
              {categories.map((cat) => (
                <div key={cat} className="role-perm-group">
                  <span className="role-perm-group__label">{cat}</span>
                  <div className="role-perm-group__items">
                    {perms.filter((p) => p.category === cat).map((p) => (
                      <label key={p.key} className={`role-perm${isAdmin ? ' is-locked' : ''}`}>
                        <input
                          type="checkbox"
                          checked={isAdmin || sel.has(p.key)}
                          disabled={isAdmin}
                          onChange={() => toggle(role.id, p.key)}
                        />
                        <span>{p.label}</span>
                      </label>
                    ))}
                  </div>
                </div>
              ))}
            </div>

            {isAdmin ? (
              <p className="muted role-card__note">{T('roles.adminNote')}</p>
            ) : (
              <div className="role-card__actions">
                <button className="btn btn--primary btn--sm" type="button" disabled={busy || !dirty(role)} onClick={() => savePerms(role)}>
                  <Save size={14} strokeWidth={2} /><span>{T('roles.savePerms')}</span>
                </button>
              </div>
            )}
          </article>
        );
      })}
    </div>
  );
}
