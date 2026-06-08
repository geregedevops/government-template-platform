"use client";

import { useCallback, useEffect, useRef, useState } from 'react';
import Link from 'next/link';
import { CheckCircle2, Loader2, ArrowLeft } from 'lucide-react';
import { useT } from '@/lib/useT';
import FormViewer from './FormViewer';
import { emptyFormSchema, type BpmRun } from '@/lib/bpm';

interface Props {
  processId: string;
}

interface ApiBody {
  ok: boolean;
  data?: BpmRun;
  message?: string;
  fieldErrors?: Record<string, string>;
}

/**
 * Процессыг ажиллуулах client. Mount дээр гүйлт эхлүүлж, идэвхтэй form
 * даалгаврыг DynamicFormRenderer-ээр зурна. Submit бүр дараагийн дэлгэц рүү
 * шилжиж, эцэст нь "дууслаа" төлөвт орно.
 */
export default function RunClient({ processId }: Props) {
  const { T } = useT();
  const [run, setRun] = useState<BpmRun | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  const start = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const res = await fetch(`/api/bpm/processes/${processId}/start`, { method: 'POST' });
      const body = (await res.json()) as ApiBody;
      if (body.ok && body.data) {
        setRun(body.data);
      } else {
        setError(body.message || T('bpm.run.error'));
      }
    } catch {
      setError(T('bpm.run.error'));
    } finally {
      setLoading(false);
    }
  }, [processId, T]);

  // Зөвхөн НЭГ удаа эхлүүлнэ (processId тус бүрд). Өмнө нь effect нь `start`-аар
  // түлхүүрлэгдсэн тул `T`-ийн identity (хэл sync) өөрчлөгдөхөд дахин ажиллаж
  // ДАВХАР instance үүсгэдэг байсан. startedForRef-ээр хамгаална.
  const startedForRef = useRef<string | null>(null);
  useEffect(() => {
    if (startedForRef.current === processId) return;
    startedForRef.current = processId;
    void start();
  }, [processId, start]);

  const submit = async (data: Record<string, unknown>) => {
    if (!run?.task) return;
    setBusy(true);
    setError('');
    try {
      const res = await fetch(`/api/bpm/tasks/${run.task.id}/submit`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ data }),
      });
      const body = (await res.json()) as ApiBody;
      if (body.ok && body.data) {
        setRun(body.data);
      } else {
        setError(body.message || T('bpm.run.error'));
      }
    } catch {
      setError(T('bpm.run.error'));
    } finally {
      setBusy(false);
    }
  };

  if (loading) {
    return (
      <div className="bpm-run__state muted">
        <Loader2 size={18} strokeWidth={2} className="spin" />
        <span>{T('bpm.run.loading')}</span>
      </div>
    );
  }

  // Эхлүүлэлт амжилтгүй (run === null) бол "дууссан" ГЭЖ ХАРУУЛАХГҮЙ — алдаа +
  // дахин оролдох товч. "Дууссан" нь зөвхөн бодит instance task-гүй болсон үед.
  if (!run) {
    return (
      <div className="bpm-run">
        {error && <div className="alert alert--danger" role="alert">{error}</div>}
        <div className="card bpm-run__state">
          <button className="btn btn--secondary" type="button" onClick={() => void start()}>
            {T('bpm.run.retry')}
          </button>
        </div>
      </div>
    );
  }

  const done = !run.task;

  return (
    <div className="bpm-run">
      {error && <div className="alert alert--danger" role="alert">{error}</div>}

      {done ? (
        <div className="card bpm-run__done">
          <CheckCircle2 size={28} strokeWidth={1.8} />
          <h2>{T('bpm.run.done')}</h2>
          <p className="muted">{T('bpm.run.doneDesc')}</p>
          <Link className="btn btn--secondary" href="/admin/bpm">
            <ArrowLeft size={16} strokeWidth={2} />
            <span>{T('bpm.run.backToList')}</span>
          </Link>
        </div>
      ) : (
        <div className="card bpm-run__step">
          {run?.task && (
            <FormViewer schema={run.task.form ?? emptyFormSchema()} busy={busy} onSubmit={submit} />
          )}
        </div>
      )}
    </div>
  );
}
