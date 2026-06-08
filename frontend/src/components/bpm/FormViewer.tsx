"use client";

import { useEffect, useRef, useState } from 'react';
import '@bpmn-io/form-js/dist/assets/form-js.css';
import { Send, Loader2 } from 'lucide-react';
import { useT } from '@/lib/useT';
import type { FormSchema } from '@/lib/bpm';
import type { FormViewerInstance } from '@bpmn-io/form-js-viewer';

interface Props {
  schema: FormSchema;
  busy?: boolean;
  onSubmit: (data: Record<string, unknown>) => void;
}

/**
 * Дэлхийн стандарт form-js (Camunda Forms) viewer-ээр user task-ийн схемийг
 * АВТОМАТААР рендерлэнэ. form-js нь модуль ачаалах үедээ DOM хэрэглэдэг тул
 * ЗӨВХӨН клиент талд динамикаар import хийнэ (SSR дээр унахгүй). createForm нь
 * нэг дуудлагаар form үүсгэж, схемийг ачаалж, рендерлэдэг албан ёсны API.
 */
export default function FormViewer({ schema, busy, onSubmit }: Props) {
  const { T } = useT();
  const hostRef = useRef<HTMLDivElement>(null);
  const formRef = useRef<FormViewerInstance | null>(null);
  const [loadError, setLoadError] = useState(false);

  useEffect(() => {
    let destroyed = false;
    let form: FormViewerInstance | null = null;

    (async () => {
      if (!hostRef.current) return;
      try {
        const { createForm } = await import('@bpmn-io/form-js-viewer');
        if (destroyed || !hostRef.current) return;
        const safe =
          schema && typeof schema === 'object' && Array.isArray((schema as { components?: unknown }).components)
            ? schema
            : { type: 'default', components: [] };
        form = await createForm({ schema: safe, data: {}, container: hostRef.current });
        formRef.current = form;
      } catch (e) {
        // eslint-disable-next-line no-console
        console.error('form-js render failed', e);
        if (!destroyed) setLoadError(true);
      }
    })();

    return () => {
      destroyed = true;
      if (form) {
        try {
          form.destroy();
        } catch {
          /* алгасах */
        }
      }
      formRef.current = null;
    };
  }, [schema]);

  const submit = () => {
    const form = formRef.current;
    if (!form || busy) return;
    const { data, errors } = form.submit();
    if (errors && Object.keys(errors).length > 0) return; // form-js inline алдаа харуулсан
    onSubmit(data);
  };

  return (
    <div className="bpm-form">
      <div ref={hostRef} className="bpm-form__host" />
      {loadError && <div className="alert alert--danger" role="alert">{T('bpm.run.formError')}</div>}
      <button className="btn btn--primary btn--lg" type="button" onClick={submit} disabled={busy}>
        {busy ? <Loader2 size={16} strokeWidth={2} className="spin" /> : <Send size={16} strokeWidth={2} />}
        <span>{busy ? T('bpm.run.submitting') : T('bpm.run.submit')}</span>
      </button>
    </div>
  );
}
