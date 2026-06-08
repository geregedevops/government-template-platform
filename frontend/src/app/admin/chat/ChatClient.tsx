"use client";

import React, { useCallback, useEffect, useRef, useState } from 'react';
import { Send, Sparkles, RotateCcw, Mic, Square, Volume2, Loader2 } from 'lucide-react';
import ReactMarkdown from 'react-markdown';
import Alert from '@/components/Alert';
import { useT } from '@/lib/useT';
import { pickMimeType, baseMime, blobToBase64, stripMarkdown, MAX_AUDIO_BYTES, MAX_RECORD_MS } from '@/lib/audio';

interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
}

/**
 * SSE event задлагч. Backend-ийн event-ууд:
 *   event: delta  data: {"delta":"..."}        — текст хэсэг
 *   event: done   data: {"conversation_id":..} — амжилттай төгсгөл
 *   event: error  data: {"message":"..."}      — алдаа (partial байж болно)
 */
function parseSSEChunk(
  buffer: string,
  onEvent: (event: string, data: string) => void,
): string {
  const frames = buffer.split('\n\n');
  const rest = frames.pop() ?? '';
  for (const frame of frames) {
    let event = 'message';
    const dataLines: string[] = [];
    for (const line of frame.split('\n')) {
      if (line.startsWith('event:')) event = line.slice(6).trim();
      else if (line.startsWith('data:')) dataLines.push(line.slice(5).trim());
    }
    if (dataLines.length > 0) onEvent(event, dataLines.join('\n'));
  }
  return rest;
}

export default function ChatClient() {
  const { T, lang } = useT();
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  // Дуу хоолойн төлвүүд.
  const [recording, setRecording] = useState(false);
  const [transcribing, setTranscribing] = useState(false);
  const [speakingIdx, setSpeakingIdx] = useState<number | null>(null);

  const conversationIdRef = useRef<string>('');
  const scrollRef = useRef<HTMLDivElement>(null);
  // Дуу бичлэг / тоглуулалтын ref-үүд.
  const recorderRef = useRef<MediaRecorder | null>(null);
  const chunksRef = useRef<Blob[]>([]);
  const streamRef = useRef<MediaStream | null>(null);
  const autoStopRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const audioElRef = useRef<HTMLAudioElement | null>(null);
  const mountedRef = useRef(true);
  const abortRef = useRef<AbortController | null>(null);

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight });
  }, [messages]);

  useEffect(() => {
    // StrictMode-д effect mount→unmount→mount хийдэг тул дахин mount болоход
    // true болгож сэргээнэ (эс бөгөөс voice бичлэг dev-д ажиллахгүй болдог).
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
      abortRef.current?.abort(); // идэвхтэй chat stream-ийг таслана
      if (autoStopRef.current) clearTimeout(autoStopRef.current);
      streamRef.current?.getTracks().forEach((t) => t.stop());
      audioElRef.current?.pause();
    };
  }, []);

  const newChat = useCallback(() => {
    conversationIdRef.current = '';
    setMessages([]);
    setError('');
  }, []);

  const send = async (e: React.FormEvent) => {
    e.preventDefault();
    const message = input.trim();
    if (!message || busy) return;

    setBusy(true);
    setError('');
    setInput('');
    setMessages((prev) => [...prev, { role: 'user', content: message }, { role: 'assistant', content: '' }]);

    const appendDelta = (delta: string) =>
      setMessages((prev) => {
        const next = [...prev];
        const last = next[next.length - 1];
        if (last?.role === 'assistant') {
          next[next.length - 1] = { ...last, content: last.content + delta };
        }
        return next;
      });

    const ctrl = new AbortController();
    abortRef.current = ctrl;
    try {
      const res = await fetch('/api/ai/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          conversation_id: conversationIdRef.current || undefined,
          message,
        }),
        signal: ctrl.signal,
      });

      const contentType = res.headers.get('content-type') ?? '';
      if (!res.ok || !contentType.includes('text/event-stream') || !res.body) {
        let msg = T('chat.error');
        try {
          const data = (await res.json()) as { message?: string };
          if (data?.message) msg = data.message;
        } catch {}
        setError(msg);
        setMessages((prev) => prev.filter((m, i) => !(i === prev.length - 1 && m.role === 'assistant' && m.content === '')));
        return;
      }

      const reader = res.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';
      let streamError = '';

      const handleEvent = (event: string, data: string) => {
        if (event === 'delta') {
          try {
            const parsed = JSON.parse(data) as { delta?: string };
            if (parsed.delta) appendDelta(parsed.delta);
          } catch {}
        } else if (event === 'done') {
          try {
            const parsed = JSON.parse(data) as { conversation_id?: string };
            if (parsed.conversation_id) conversationIdRef.current = parsed.conversation_id;
          } catch {}
        } else if (event === 'error') {
          try {
            const parsed = JSON.parse(data) as { message?: string };
            streamError = parsed.message ?? T('chat.error');
          } catch {
            streamError = T('chat.error');
          }
        }
      };

      // eslint-disable-next-line no-constant-condition
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        buffer += decoder.decode(value, { stream: true });
        buffer = parseSSEChunk(buffer, handleEvent);
      }
      // EOF flush: сүүлийн event нь `\n\n`-ээр төгсөөгүй бол buffer-т үлдсэн
      // байж болзошгүй (done/conversation_id алдагдахаас сэргийлнэ).
      buffer += decoder.decode();
      if (buffer.trim()) {
        parseSSEChunk(buffer.endsWith('\n\n') ? buffer : buffer + '\n\n', handleEvent);
      }

      if (streamError) setError(streamError);
      setMessages((prev) => prev.filter((m, i) => !(i === prev.length - 1 && m.role === 'assistant' && m.content === '')));
    } catch (e) {
      // Зориудаар тассан (unmount/navigate) бол алдаа харуулахгүй.
      if (!(e instanceof DOMException && e.name === 'AbortError')) {
        setError(T('chat.networkError'));
        setMessages((prev) => prev.filter((m, i) => !(i === prev.length - 1 && m.role === 'assistant' && m.content === '')));
      }
    } finally {
      if (abortRef.current === ctrl) abortRef.current = null;
      setBusy(false);
    }
  };

  // --- STT: дуугаар асуух (бичлэг → бичвэр → input) ---

  const transcribe = useCallback(
    async (blob: Blob, mimeType: string) => {
      if (blob.size > MAX_AUDIO_BYTES) {
        setError(T('translate.tooLong'));
        return;
      }
      setTranscribing(true);
      setError('');
      try {
        const audioBase64 = await blobToBase64(blob);
        const res = await fetch('/api/voice/transcribe', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ lang, mime_type: baseMime(mimeType), audio_base64: audioBase64 }),
        });
        const payload = (await res.json()) as { ok?: boolean; data?: { text?: string }; message?: string };
        if (!res.ok || !payload.ok || !payload.data?.text) {
          setError(payload.message || T('chat.voiceError'));
          return;
        }
        // Танисан бичвэрийг оролтод нэмнэ — хэрэглэгч засаад илгээж болно.
        setInput((prev) => (prev ? `${prev} ${payload.data!.text}` : payload.data!.text!));
      } catch {
        setError(T('chat.networkError'));
      } finally {
        setTranscribing(false);
      }
    },
    [lang, T],
  );

  const stopRecording = useCallback(() => {
    if (autoStopRef.current) {
      clearTimeout(autoStopRef.current);
      autoStopRef.current = null;
    }
    if (recorderRef.current?.state === 'recording') recorderRef.current.stop();
  }, []);

  const startRecording = useCallback(async () => {
    setError('');
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
      if (blob.size > 0) void transcribe(blob, effectiveMime);
    };
    recorderRef.current = recorder;
    recorder.start();
    setRecording(true);
    autoStopRef.current = setTimeout(() => stopRecording(), MAX_RECORD_MS);
  }, [T, transcribe, stopRecording]);

  const toggleRecord = () => (recording ? stopRecording() : startRecording());

  // --- TTS: хариуг чанга унших ---

  const speak = useCallback(
    async (idx: number, text: string) => {
      // Дахин дарвал зогсооно.
      if (speakingIdx === idx) {
        audioElRef.current?.pause();
        audioElRef.current = null;
        setSpeakingIdx(null);
        return;
      }
      audioElRef.current?.pause();
      setSpeakingIdx(idx);
      setError('');
      try {
        const res = await fetch('/api/voice/speak', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          // Markdown тэмдэгтийг хасч, богиносгож илгээнэ — TTS-ийн хугацааны
          // төсөвт багтааж, цэвэр яриа гаргана.
          body: JSON.stringify({ text: stripMarkdown(text) }),
        });
        const payload = (await res.json()) as { ok?: boolean; data?: { audio_base64?: string; audio_mime?: string }; message?: string };
        if (!res.ok || !payload.ok || !payload.data?.audio_base64) {
          // 5xx-ийн ерөнхий мессежийн оронд TTS-д тодорхой мессеж харуулна.
          setError(res.status >= 500 ? T('chat.ttsError') : payload.message || T('chat.ttsError'));
          setSpeakingIdx(null);
          return;
        }
        const audio = new Audio(`data:${payload.data.audio_mime || 'audio/wav'};base64,${payload.data.audio_base64}`);
        audioElRef.current = audio;
        audio.onended = () => {
          if (audioElRef.current === audio) audioElRef.current = null;
          setSpeakingIdx((cur) => (cur === idx ? null : cur));
        };
        // Аудио ачаалах/тоглуулах алдаа гарвал spinner гацахаас сэргийлнэ.
        audio.onerror = () => {
          if (audioElRef.current === audio) audioElRef.current = null;
          setError(T('chat.ttsError'));
          setSpeakingIdx((cur) => (cur === idx ? null : cur));
        };
        await audio.play();
      } catch {
        setError(T('chat.networkError'));
        setSpeakingIdx(null);
      }
    },
    [speakingIdx, T],
  );

  return (
    <section className="card" aria-label={T('chat.title')} style={{ display: 'flex', flexDirection: 'column', minHeight: 420 }}>
      <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 8 }}>
        <button className="btn btn--secondary" type="button" onClick={newChat} disabled={busy}>
          <RotateCcw size={14} strokeWidth={2} />
          <span>{T('chat.newChat')}</span>
        </button>
      </div>

      <div
        ref={scrollRef}
        style={{ flex: 1, overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: 12, padding: '4px 2px', maxHeight: 480 }}
        aria-live="polite"
      >
        {messages.length === 0 && (
          <p style={{ color: 'var(--muted)', fontSize: 13, display: 'flex', alignItems: 'center', gap: 8 }}>
            <Sparkles size={14} strokeWidth={2} style={{ color: 'var(--dan-blue-text)' }} />
            {T('chat.empty')}
          </p>
        )}
        {messages.map((m, i) => (
          <div
            key={i}
            style={{
              alignSelf: m.role === 'user' ? 'flex-end' : 'flex-start',
              maxWidth: '85%',
              borderRadius: 12,
              padding: '10px 14px',
              fontSize: 14,
              lineHeight: 1.55,
              // Хэрэглэгчийн мессеж энгийн текст (мөр хадгална); туслахынх нь
              // markdown болж render хийгдэнэ.
              whiteSpace: m.role === 'user' ? 'pre-wrap' : 'normal',
              background: m.role === 'user' ? 'var(--dan-blue-soft)' : 'var(--card-2, var(--bg-2, rgba(127,127,127,0.08)))',
              border: '1px solid var(--border)',
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 8, marginBottom: 4 }}>
              <span style={{ fontSize: 11, fontWeight: 600, color: 'var(--muted)' }}>
                {m.role === 'user' ? T('chat.you') : T('chat.assistant')}
              </span>
              {m.role === 'assistant' && m.content && (
                <button
                  type="button"
                  onClick={() => speak(i, m.content)}
                  title={speakingIdx === i ? T('chat.speaking') : T('chat.listen')}
                  aria-label={speakingIdx === i ? T('chat.speaking') : T('chat.listen')}
                  style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--dan-blue-text)', padding: 2, display: 'inline-flex' }}
                >
                  {speakingIdx === i ? <Loader2 size={14} className="spin" strokeWidth={2} /> : <Volume2 size={14} strokeWidth={2} />}
                </button>
              )}
            </div>
            {m.role === 'assistant'
              ? (m.content
                  ? <div className="chat-md"><ReactMarkdown>{m.content}</ReactMarkdown></div>
                  : (busy && i === messages.length - 1 ? T('chat.thinking') : ''))
              : m.content}
          </div>
        ))}
      </div>

      {error && <div style={{ marginTop: 10 }}><Alert kind="danger">{error}</Alert></div>}

      <form onSubmit={send} style={{ display: 'flex', gap: 8, marginTop: 12 }}>
        <button
          type="button"
          onClick={toggleRecord}
          disabled={busy || transcribing}
          aria-pressed={recording}
          title={recording ? T('chat.listening') : T('chat.record')}
          aria-label={recording ? T('chat.listening') : T('chat.record')}
          className={`btn ${recording ? 'btn--danger' : 'btn--secondary'}`}
        >
          {transcribing ? <Loader2 size={16} className="spin" strokeWidth={2} /> : recording ? <Square size={16} strokeWidth={2} /> : <Mic size={16} strokeWidth={2} />}
        </button>
        <input
          className="input"
          style={{ flex: 1 }}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder={recording ? T('chat.listening') : transcribing ? T('chat.transcribing') : T('chat.placeholder')}
          maxLength={4000}
          disabled={busy || recording || transcribing}
          aria-label={T('chat.placeholder')}
        />
        <button className="btn btn--primary" type="submit" disabled={busy || recording || transcribing || input.trim() === ''}>
          <Send size={16} strokeWidth={2} />
          <span>{busy ? T('chat.thinking') : T('chat.send')}</span>
        </button>
      </form>
    </section>
  );
}
