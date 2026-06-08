"use client";

import { useCallback, useEffect, useState } from 'react';
import { Loader2, Activity, ChevronDown, ChevronRight } from 'lucide-react';
import { useT } from '@/lib/useT';
import type { DictKey } from '@/lib/i18n';
import type { BpmInstance, BpmEvent } from '@/lib/bpm';

interface Props {
  processId: string;
}

interface ListResponse {
  ok: boolean;
  data?: BpmInstance[];
  message?: string;
}
interface EventsResponse {
  ok: boolean;
  data?: BpmEvent[];
}

export default function InstanceList({ processId }: Props) {
  const { T, lang } = useT();
  const [items, setItems] = useState<BpmInstance[] | null>(null);
  const [error, setError] = useState('');
  const [openId, setOpenId] = useState<string | null>(null);
  // instanceId -> events ('loading' эсвэл массив).
  const [events, setEvents] = useState<Record<string, BpmEvent[] | 'loading'>>({});

  const load = useCallback(async () => {
    setError('');
    try {
      const res = await fetch(`/api/bpm/processes/${processId}/instances`, { method: 'GET' });
      const body = (await res.json()) as ListResponse;
      if (body.ok) setItems(body.data ?? []);
      else {
        setItems([]);
        setError(body.message || T('bpm.runs.loadError'));
      }
    } catch {
      setItems([]);
      setError(T('bpm.runs.loadError'));
    }
  }, [processId, T]);

  useEffect(() => {
    void load();
  }, [load]);

  const toggle = async (id: string) => {
    if (openId === id) {
      setOpenId(null);
      return;
    }
    setOpenId(id);
    if (!events[id]) {
      setEvents((prev) => ({ ...prev, [id]: 'loading' }));
      try {
        const res = await fetch(`/api/bpm/instances/${id}/events`, { method: 'GET' });
        const body = (await res.json()) as EventsResponse;
        setEvents((prev) => ({ ...prev, [id]: body.ok ? body.data ?? [] : [] }));
      } catch {
        setEvents((prev) => ({ ...prev, [id]: [] }));
      }
    }
  };

  return (
    <div className="bpm-runs">
      {error && <div className="alert alert--danger" role="alert">{error}</div>}

      {items === null && (
        <div className="bpm-list__loading muted">
          <Loader2 size={16} strokeWidth={2} className="spin" />
          <span>{T('bpm.run.loading')}</span>
        </div>
      )}

      {items !== null && items.length === 0 && !error && (
        <div className="card bpm-empty">
          <Activity size={22} strokeWidth={1.6} />
          <p className="muted">{T('bpm.runs.empty')}</p>
        </div>
      )}

      {items !== null && items.length > 0 && (
        <div className="bpm-runs__list">
          {items.map((inst) => {
            const open = openId === inst.id;
            const evs = events[inst.id];
            const hasVars = inst.variables && Object.keys(inst.variables).length > 0;
            return (
              <article key={inst.id} className="card bpm-run-row">
                <div className="bpm-run-row__main">
                  <span className={`bpm-badge bpm-badge--${inst.status}`}>{T(statusKey(inst.status))}</span>
                  <div className="bpm-run-row__meta">
                    <span className="defrow__label">{T('bpm.runs.started')}</span>
                    <span className="mono">{formatDateTime(inst.created_at, lang)}</span>
                  </div>
                  <div className="bpm-run-row__meta">
                    <span className="defrow__label">{T('bpm.runs.finished')}</span>
                    <span className="mono">{inst.completed_at ? formatDateTime(inst.completed_at, lang) : '—'}</span>
                  </div>
                  <div className="bpm-run-row__meta">
                    <span className="defrow__label">{T('bpm.runs.step')}</span>
                    <span className="mono">{inst.current_node_id || '—'}</span>
                  </div>
                  <button className="btn btn--ghost btn--sm" type="button" onClick={() => toggle(inst.id)}>
                    {open ? <ChevronDown size={14} strokeWidth={2} /> : <ChevronRight size={14} strokeWidth={2} />}
                    <span>{T('bpm.runs.timeline')}</span>
                  </button>
                </div>

                {open && (
                  <div className="bpm-run-row__detail">
                    {evs === 'loading' && (
                      <div className="muted bpm-list__loading">
                        <Loader2 size={14} strokeWidth={2} className="spin" />
                        <span>{T('bpm.run.loading')}</span>
                      </div>
                    )}
                    {Array.isArray(evs) && evs.length > 0 && (
                      <ol className="bpm-timeline">
                        {evs.map((e) => (
                          <li key={e.id} className={`bpm-timeline__item bpm-timeline__item--${e.type}`}>
                            <span className="bpm-timeline__dot" aria-hidden="true" />
                            <div className="bpm-timeline__body">
                              <span className="bpm-timeline__label">{T(eventKey(e.type))}</span>
                              {e.node_id && <span className="chip mono">{e.node_id}</span>}
                              {e.detail && <span className="muted bpm-timeline__detail">{e.detail}</span>}
                            </div>
                            <span className="muted mono bpm-timeline__time">{formatTime(e.created_at, lang)}</span>
                          </li>
                        ))}
                      </ol>
                    )}
                    {Array.isArray(evs) && evs.length === 0 && (
                      <p className="muted" style={{ fontSize: 13 }}>{T('bpm.runs.noEvents')}</p>
                    )}
                    {hasVars && (
                      <details className="bpm-run-row__vars-wrap">
                        <summary className="muted">{T('bpm.runs.variables')}</summary>
                        <pre className="bpm-run-row__vars mono">{JSON.stringify(inst.variables, null, 2)}</pre>
                      </details>
                    )}
                  </div>
                )}
              </article>
            );
          })}
        </div>
      )}
    </div>
  );
}

function statusKey(status: string): DictKey {
  switch (status) {
    case 'completed':
      return 'bpm.inst.completed';
    case 'failed':
      return 'bpm.inst.failed';
    case 'cancelled':
      return 'bpm.inst.cancelled';
    default:
      return 'bpm.inst.running';
  }
}

function eventKey(type: string): DictKey {
  const known: Record<string, DictKey> = {
    instance_started: 'bpm.evt.instance_started',
    task_opened: 'bpm.evt.task_opened',
    task_completed: 'bpm.evt.task_completed',
    service_called: 'bpm.evt.service_called',
    service_failed: 'bpm.evt.service_failed',
    gateway_routed: 'bpm.evt.gateway_routed',
    instance_completed: 'bpm.evt.instance_completed',
    instance_failed: 'bpm.evt.instance_failed',
  };
  return known[type] ?? 'bpm.evt.unknown';
}

function formatDateTime(iso: string, lang: string): string {
  try {
    return new Date(iso).toLocaleString(lang === 'en' ? 'en-US' : 'mn-MN', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return iso;
  }
}

function formatTime(iso: string, lang: string): string {
  try {
    return new Date(iso).toLocaleTimeString(lang === 'en' ? 'en-US' : 'mn-MN', {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  } catch {
    return iso;
  }
}
