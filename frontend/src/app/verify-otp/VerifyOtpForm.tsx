"use client";

import React, { useEffect, useRef, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { ShieldCheck, Send, ArrowLeft } from 'lucide-react';
import Alert from '@/components/Alert';
import { postJSON } from '@/lib/client';
import { useT } from '@/lib/useT';

export default function VerifyOtpForm({
  initialEmail, autosend,
}: {
  initialEmail: string;
  autosend: boolean;
}) {
  const { T } = useT();
  const router = useRouter();
  const [email, setEmail] = useState(initialEmail);
  const [code, setCode] = useState('');
  const [busy, setBusy] = useState(false);
  const [resending, setResending] = useState(false);
  const [error, setError] = useState('');
  const [info, setInfo] = useState('');
  const didAutosend = useRef(false);

  const sendCode = async () => {
    if (!email) {
      setError(T('auth.otp.needEmail'));
      return;
    }
    setResending(true);
    setError('');
    const res = await postJSON('/api/auth/send-otp', { email });
    setResending(false);
    if (res.ok) {
      setInfo(T('auth.otp.sent'));
    } else {
      setError(res.message ?? T('auth.otp.sendFailed'));
    }
  };

  // Бүртгэлийн дараа орж ирэхэд кодыг нэг л удаа автоматаар илгээнэ.
  useEffect(() => {
    if (autosend && initialEmail && !didAutosend.current) {
      didAutosend.current = true;
      void sendCode();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError('');
    const res = await postJSON('/api/auth/verify-otp', { email, code });
    setBusy(false);
    if (res.ok) {
      router.push('/login?notice=verified');
      return;
    }
    setError(res.message ?? T('auth.otp.verifyFailed'));
  };

  return (
    <form className="form-grid" onSubmit={submit} noValidate>
      {info && <Alert kind="info">{info}</Alert>}
      {error && <Alert kind="danger">{error}</Alert>}

      <div className="field">
        <label className="field__label" htmlFor="email">{T('auth.email')}</label>
        <input
          id="email"
          name="email"
          type="email"
          className="input"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          autoComplete="email"
          placeholder="tani@gerege.mn"
          readOnly={!!initialEmail}
          required
        />
      </div>

      <div className="field">
        <label className="field__label" htmlFor="code">{T('auth.otp.code')}</label>
        <input
          id="code"
          name="code"
          className="input otp-input"
          value={code}
          onChange={(e) => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
          inputMode="numeric"
          autoComplete="one-time-code"
          placeholder="000000"
          maxLength={6}
          required
        />
        <span className="field__help">{T('auth.otp.codeHelp')}</span>
      </div>

      <button className="btn btn--primary btn--lg btn--block" type="submit" disabled={busy || code.length < 6}>
        <ShieldCheck size={18} strokeWidth={2} />
        <span>{busy ? T('auth.otp.verifying') : T('auth.otp.verify')}</span>
      </button>

      <button className="btn btn--secondary btn--block" type="button" onClick={sendCode} disabled={resending}>
        <Send size={16} strokeWidth={2} />
        <span>{resending ? T('auth.otp.resending') : T('auth.otp.resend')}</span>
      </button>

      <p className="signin-card__alt">
        <Link href="/login" style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}><ArrowLeft size={14} strokeWidth={2} />{T('auth.backToLogin')}</Link>
      </p>
    </form>
  );
}
