"use client";

import React, { useCallback, useEffect, useRef, useState } from 'react';
import { Mic, Square, ArrowLeftRight, Volume2, Loader2 } from 'lucide-react';
import Alert from '@/components/Alert';
import { useT } from '@/lib/useT';
import { pickMimeType, baseMime, blobToBase64, MAX_AUDIO_BYTES, MAX_RECORD_MS } from '@/lib/audio';

type Lang = 'mn' | 'en';

interface TranslateResult {
  id: string;
  source_lang: Lang;
  target_lang: Lang;
  source_text: string;
  translated_text: string;
  audio_base64: string;
  audio_mime: string;
}

interface HistoryItem {
  id: string;
  source_lang: Lang;
  target_lang: Lang;
  source_text: string;
  translated_text: string;
  created_at: string;
}

export default function TranslateClient() {
  const { T, lang: uiLang } = useT();
  const [source, setSource] = useState<Lang>('mn');
  const [recording, setRecording] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [result, setResult] = useState<TranslateResult | null>(null);
  const [history, setHistory] = useState<HistoryItem[]>([]);

  const recorderRef = useRef<MediaRecorder | null>(null);
  const chunksRef = useRef<Blob[]>([]);
  const streamRef = useRef<MediaStream | null>(null);
  const autoStopRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const mountedRef = useRef(true);

  const target: Lang = source === 'mn' ? 'en' : 'mn';

  const loadHistory = useCallback(async () => {
    try {
      const res = await fetch('/api/voice/history', { cache: 'no-store' });
      const data = (await res.json()) as { ok?: boolean; data?: HistoryItem[] };
      if (data.ok && Array.isArray(data.data)) setHistory(data.data);
    } catch {
      /* түүх ачаалж чадсангүй — чухал биш */
    }
  }, []);

  useEffect(() => {
    // StrictMode-д дахин mount болоход true болгож сэргээнэ.
    mountedRef.current = true;
    loadHistory();
    return () => {
      mountedRef.current = false;
      if (autoStopRef.current) clearTimeout(autoStopRef.current);
      streamRef.current?.getTracks().forEach((t) => t.stop());
    };
  }, [loadHistory]);

  const sendAudio = useCallback(
    async (blob: Blob, mimeType: string) => {
      if (blob.size > MAX_AUDIO_BYTES) {
        setError(T('translate.tooLong'));
        return;
      }
      setBusy(true);
      setError('');
      try {
        const audioBase64 = await blobToBase64(blob);
        const res = await fetch('/api/voice/translate', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            source_lang: source,
            mime_type: baseMime(mimeType),
            audio_base64: audioBase64,
          }),
        });
        const payload = (await res.json()) as { ok?: boolean; data?: TranslateResult; message?: string };
        if (!res.ok || !payload.ok || !payload.data) {
          setError(payload.message || T('translate.error'));
          return;
        }
        setResult(payload.data);
        loadHistory();
      } catch {
        setError(T('translate.networkError'));
      } finally {
        setBusy(false);
      }
    },
    [source, T, loadHistory],
  );

  const stopRecording = useCallback(() => {
    if (autoStopRef.current) {
      clearTimeout(autoStopRef.current);
      autoStopRef.current = null;
    }
    recorderRef.current?.state === 'recording' && recorderRef.current.stop();
  }, []);

  const startRecording = useCallback(async () => {
    setError('');
    setResult(null);
    if (typeof MediaRecorder === 'undefined' || !navigator.mediaDevices?.getUserMedia) {
      setError(T('translate.unsupported'));
      return;
    }
    let stream: MediaStream;
    try {
      stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    } catch {
      setError(T('translate.micDenied'));
      return;
    }
    // Зөвшөөрлийн харилцах цонх нээлттэй байх зуур хэрэглэгч хуудаснаас
    // гарвал (unmount) олж авсан stream-ийг шууд хааж, орхигдсон микрофон
    // нээлттэй үлдэхээс сэргийлнэ.
    if (!mountedRef.current) {
      stream.getTracks().forEach((t) => t.stop());
      return;
    }
    streamRef.current = stream;
    const mimeType = pickMimeType();
    const recorder = mimeType ? new MediaRecorder(stream, { mimeType }) : new MediaRecorder(stream);
    chunksRef.current = [];

    recorder.ondataavailable = (e) => {
      if (e.data.size > 0) chunksRef.current.push(e.data);
    };
    recorder.onstop = () => {
      stream.getTracks().forEach((t) => t.stop());
      streamRef.current = null;
      setRecording(false);
      const effectiveMime = recorder.mimeType || mimeType || 'audio/webm';
      const blob = new Blob(chunksRef.current, { type: effectiveMime });
      if (blob.size > 0) void sendAudio(blob, effectiveMime);
    };

    recorderRef.current = recorder;
    recorder.start();
    setRecording(true);
    // Авто-зогсолт — урт бичлэгээс сэргийлнэ.
    autoStopRef.current = setTimeout(() => stopRecording(), MAX_RECORD_MS);
  }, [T, sendAudio, stopRecording]);

  const toggleRecord = () => (recording ? stopRecording() : startRecording());
  const langLabel = (l: Lang) => (l === 'mn' ? (uiLang === 'en' ? 'Mongolian' : 'Монгол') : uiLang === 'en' ? 'English' : 'Англи');

  return (
    <section className="card" aria-label={T('translate.title')} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      {/* Чиглэл сонгогч */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 12 }}>
        <span style={{ fontWeight: 600, fontSize: 14 }}>{langLabel(source)}</span>
        <button
          className="btn btn--secondary"
          type="button"
          onClick={() => setSource((s) => (s === 'mn' ? 'en' : 'mn'))}
          disabled={recording || busy}
          aria-label={T('translate.swap')}
          title={T('translate.swap')}
        >
          <ArrowLeftRight size={16} strokeWidth={2} />
        </button>
        <span style={{ fontWeight: 600, fontSize: 14 }}>{langLabel(target)}</span>
      </div>

      {/* Бичлэгийн товч */}
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 8 }}>
        <button
          type="button"
          onClick={toggleRecord}
          disabled={busy}
          aria-pressed={recording}
          className={`btn ${recording ? 'btn--danger' : 'btn--primary'}`}
          style={{ width: 96, height: 96, borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
        >
          {busy ? <Loader2 size={32} className="spin" strokeWidth={2} /> : recording ? <Square size={30} strokeWidth={2} /> : <Mic size={32} strokeWidth={2} />}
        </button>
        <span style={{ fontSize: 13, color: 'var(--muted)' }}>
          {busy ? T('translate.processing') : recording ? T('translate.recording') : result ? '' : T('translate.empty')}
        </span>
        {!busy && !recording && (
          <span style={{ fontSize: 12, color: 'var(--muted)' }}>{recording ? '' : T('translate.record')}</span>
        )}
      </div>

      {error && <Alert kind="danger">{error}</Alert>}

      {/* Үр дүн */}
      {result && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div style={{ borderRadius: 12, padding: '12px 14px', border: '1px solid var(--border)', background: 'var(--card-2, rgba(127,127,127,0.06))' }}>
            <div style={{ fontSize: 11, fontWeight: 600, color: 'var(--muted)', marginBottom: 4 }}>
              {T('translate.source')} · {langLabel(result.source_lang)}
            </div>
            <div style={{ fontSize: 15, lineHeight: 1.5 }}>{result.source_text || '—'}</div>
          </div>
          <div style={{ borderRadius: 12, padding: '12px 14px', border: '1px solid var(--border)', background: 'var(--dan-blue-soft)' }}>
            <div style={{ fontSize: 11, fontWeight: 600, color: 'var(--muted)', marginBottom: 4 }}>
              {T('translate.translation')} · {langLabel(result.target_lang)}
            </div>
            <div style={{ fontSize: 16, fontWeight: 500, lineHeight: 1.5 }}>{result.translated_text || '—'}</div>
            {result.audio_base64 && (
              <div style={{ marginTop: 10, display: 'flex', alignItems: 'center', gap: 8 }}>
                <Volume2 size={16} strokeWidth={2} style={{ color: 'var(--dan-blue-text)' }} />
                {/* key нь шинэ үр дүн бүрт audio элементийг дахин ачаалж autoPlay-г идэвхжүүлнэ */}
                <audio
                  key={result.id}
                  controls
                  autoPlay
                  src={`data:${result.audio_mime};base64,${result.audio_base64}`}
                  style={{ height: 36, maxWidth: '100%' }}
                >
                  <track kind="captions" />
                </audio>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Түүх */}
      {history.length > 0 && (
        <div style={{ marginTop: 4 }}>
          <div style={{ fontSize: 12, fontWeight: 600, color: 'var(--muted)', marginBottom: 8 }}>{T('translate.history')}</div>
          <ul style={{ listStyle: 'none', margin: 0, padding: 0, display: 'flex', flexDirection: 'column', gap: 8 }}>
            {history.slice(0, 8).map((h) => (
              <li key={h.id} style={{ fontSize: 13, borderBottom: '1px solid var(--border)', paddingBottom: 8 }}>
                <span style={{ color: 'var(--muted)' }}>{langLabel(h.source_lang)} → {langLabel(h.target_lang)}: </span>
                <span>{h.source_text}</span>
                <span style={{ color: 'var(--dan-blue-text)' }}> → {h.translated_text}</span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </section>
  );
}
