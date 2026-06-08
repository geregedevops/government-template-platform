"use client";

import { useCallback, useEffect, useState } from 'react';
import { Plus, Trash2, Loader2, Users as UsersIcon, X, Save } from 'lucide-react';
import { useT } from '@/lib/useT';

const ROLE_USER = 2; // шинэ хэрэглэгчийн өгөгдмөл эрх

interface AdminUser {
  id: string;
  username: string;
  email: string;
  role_id: number;
  org_id?: string;
  created_at: string;
}
interface ApiResponse<T> {
  ok: boolean;
  data?: T;
  message?: string;
}

interface Props {
  currentUserId: string;
}

export default function UsersManager({ currentUserId }: Props) {
  const { T, lang } = useT();
  const [items, setItems] = useState<AdminUser[] | null>(null);
  const [error, setError] = useState('');
  const [adding, setAdding] = useState(false);
  const [busy, setBusy] = useState(false);
  // шинэ хэрэглэгчийн форм
  const [form, setForm] = useState({ username: '', email: '', password: '', role_id: ROLE_USER });
  // бүх эрх (role) — dropdown-д ашиглана (динамик RBAC).
  const [roles, setRoles] = useState<{ id: number; name: string }[]>([]);
  // байгууллагууд — хэрэглэгчийг хуваарилах dropdown-д.
  const [orgs, setOrgs] = useState<{ id: string; name: string }[]>([]);

  const load = useCallback(async () => {
    setError('');
    try {
      const [uRes, rRes, oRes] = await Promise.all([
        fetch('/api/users', { method: 'GET' }),
        fetch('/api/rbac/roles', { method: 'GET' }),
        fetch('/api/orgs', { method: 'GET' }),
      ]);
      const body = (await uRes.json()) as ApiResponse<AdminUser[]>;
      const rBody = (await rRes.json()) as ApiResponse<{ id: number; name: string }[]>;
      const oBody = (await oRes.json()) as ApiResponse<{ id: string; name: string }[]>;
      if (rBody.ok) setRoles(rBody.data ?? []);
      if (oBody.ok) setOrgs(oBody.data ?? []);
      if (body.ok) setItems(body.data ?? []);
      else {
        setItems([]);
        setError(body.message || T('users.loadError'));
      }
    } catch {
      setItems([]);
      setError(T('users.loadError'));
    }
  }, [T]);

  useEffect(() => {
    void load();
  }, [load]);

  const create = async () => {
    if (!form.username.trim() || !form.email.trim() || form.password.length < 8) {
      setError(T('users.createInvalid'));
      return;
    }
    setBusy(true);
    setError('');
    try {
      const res = await fetch('/api/users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(form),
      });
      const body = (await res.json()) as ApiResponse<AdminUser>;
      setBusy(false);
      if (body.ok) {
        setAdding(false);
        setForm({ username: '', email: '', password: '', role_id: ROLE_USER });
        void load();
        return;
      }
      setError(body.message || T('users.saveError'));
    } catch {
      setBusy(false);
      setError(T('users.saveError'));
    }
  };

  const changeRole = async (id: string, roleId: number) => {
    setError('');
    try {
      const res = await fetch(`/api/users/${id}/role`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ role_id: roleId }),
      });
      const body = (await res.json()) as ApiResponse<unknown>;
      if (body.ok) void load();
      else setError(body.message || T('users.saveError'));
    } catch {
      setError(T('users.saveError'));
    }
  };

  const changeOrg = async (id: string, orgId: string) => {
    setError('');
    try {
      const res = await fetch(`/api/users/${id}/org`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ org_id: orgId }),
      });
      const body = (await res.json()) as ApiResponse<unknown>;
      if (body.ok) void load();
      else setError(body.message || T('users.saveError'));
    } catch {
      setError(T('users.saveError'));
    }
  };

  const remove = async (id: string) => {
    if (!window.confirm(T('users.deleteConfirm'))) return;
    setError('');
    try {
      const res = await fetch(`/api/users/${id}`, { method: 'DELETE' });
      const body = (await res.json()) as ApiResponse<unknown>;
      if (body.ok) void load();
      else setError(body.message || T('users.deleteError'));
    } catch {
      setError(T('users.deleteError'));
    }
  };

  const fmtDate = (iso: string) => {
    try {
      return new Date(iso).toLocaleDateString(lang === 'en' ? 'en-US' : 'mn-MN', {
        year: 'numeric', month: 'short', day: 'numeric',
      });
    } catch {
      return iso;
    }
  };

  return (
    <div className="users">
      <div className="users__head">
        {!adding && (
          <button className="btn btn--primary" type="button" onClick={() => { setAdding(true); setError(''); }}>
            <Plus size={16} strokeWidth={2} />
            <span>{T('users.add')}</span>
          </button>
        )}
      </div>

      {error && <div className="alert alert--danger" role="alert">{error}</div>}

      {adding && (
        <div className="card users-form">
          <div className="users-form__head">
            <UsersIcon size={15} strokeWidth={2} />
            <span>{T('users.add')}</span>
          </div>
          <div className="users-form__grid">
            <div className="field">
              <label className="field__label" htmlFor="u-name">{T('users.username')}</label>
              <input id="u-name" className="input" value={form.username} onChange={(e) => setForm({ ...form, username: e.target.value })} />
            </div>
            <div className="field">
              <label className="field__label" htmlFor="u-email">{T('users.email')}</label>
              <input id="u-email" className="input" type="email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} />
            </div>
            <div className="field">
              <label className="field__label" htmlFor="u-pass">{T('users.password')}</label>
              <input id="u-pass" className="input" type="password" value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} />
            </div>
            <div className="field">
              <label className="field__label" htmlFor="u-role">{T('users.role')}</label>
              <select id="u-role" className="input" value={form.role_id} onChange={(e) => setForm({ ...form, role_id: Number(e.target.value) })}>
                {roles.map((r) => <option key={r.id} value={r.id}>{r.name}</option>)}
              </select>
            </div>
          </div>
          <div className="users-form__actions">
            <button className="btn btn--primary" type="button" onClick={create} disabled={busy}>
              {busy ? <Loader2 size={16} strokeWidth={2} className="spin" /> : <Save size={16} strokeWidth={2} />}
              <span>{T('users.create')}</span>
            </button>
            <button className="btn btn--secondary" type="button" onClick={() => setAdding(false)} disabled={busy}>
              <X size={16} strokeWidth={2} />
              <span>{T('users.cancel')}</span>
            </button>
          </div>
        </div>
      )}

      {items === null && (
        <div className="bpm-list__loading muted">
          <Loader2 size={16} strokeWidth={2} className="spin" />
          <span>{T('users.loading')}</span>
        </div>
      )}

      {items !== null && items.length > 0 && (
        <div className="card users-table-wrap">
          <table className="users-table">
            <thead>
              <tr>
                <th>{T('users.username')}</th>
                <th>{T('users.email')}</th>
                <th>{T('users.role')}</th>
                <th>{T('users.org')}</th>
                <th>{T('users.created')}</th>
                <th aria-label="actions" />
              </tr>
            </thead>
            <tbody>
              {items.map((u) => {
                const isSelf = u.id === currentUserId;
                return (
                  <tr key={u.id}>
                    <td>{u.username}{isSelf && <span className="chip chip--neutral" style={{ marginLeft: 8 }}>{T('users.you')}</span>}</td>
                    <td className="mono">{u.email}</td>
                    <td>
                      <select
                        className="input users-table__role"
                        value={u.role_id}
                        disabled={isSelf}
                        onChange={(e) => changeRole(u.id, Number(e.target.value))}
                      >
                        {roles.map((r) => <option key={r.id} value={r.id}>{r.name}</option>)}
                      </select>
                    </td>
                    <td>
                      <select
                        className="input users-table__role"
                        value={u.org_id ?? ''}
                        onChange={(e) => changeOrg(u.id, e.target.value)}
                      >
                        {orgs.map((o) => <option key={o.id} value={o.id}>{o.name}</option>)}
                      </select>
                    </td>
                    <td className="mono">{fmtDate(u.created_at)}</td>
                    <td className="users-table__actions">
                      {!isSelf && (
                        <button className="btn btn--ghost btn--sm" type="button" onClick={() => remove(u.id)} aria-label={T('users.delete')}>
                          <Trash2 size={14} strokeWidth={2} />
                        </button>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
