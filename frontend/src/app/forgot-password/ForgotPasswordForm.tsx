"use client";

import React, { useState } from 'react';
import Link from 'next/link';
import { Mail, ArrowLeft } from 'lucide-react';
import Alert from '@/components/Alert';
import { postJSON } from '@/lib/client';
import { useT } from '@/lib/useT';

export default function ForgotPasswordForm() {
  const { T } = useT();
  const [email, setEmail] = useState('');
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState('');

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError('');
    const res = await postJSON('/api/auth/forgot-password', { email });
    setBusy(false);
    // Backend нь enumeration-аас сэргийлж бүртгэлтэй эсэхээс үл хамааран ижил
    // хариу буцаадаг — тиймээс амжилттай үед үргэлж ерөнхий мессеж харуулна.
    if (res.ok || res.status === 200) {
      setDone(true);
    } else if (res.status === 422 && res.fieldErrors?.email) {
      setError(res.fieldErrors.email);
    } else {
      setError(res.message ?? T('auth.forgot.failed'));
    }
  };

  if (done) {
    return (
      <div className="form-grid">
        <Alert kind="success">{T('auth.forgot.sent')}</Alert>
        <p className="signin-card__alt">
          <Link href="/login" style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}><ArrowLeft size={14} strokeWidth={2} />{T('auth.backToLogin')}</Link>
        </p>
      </div>
    );
  }

  return (
    <form className="form-grid" onSubmit={submit} noValidate>
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
          required
        />
        <span className="field__help">{T('auth.email.helpRegistered')}</span>
      </div>

      <button className="btn btn--primary btn--lg btn--block" type="submit" disabled={busy}>
        <Mail size={18} strokeWidth={2} />
        <span>{busy ? T('auth.forgot.submitting') : T('auth.forgot.submit')}</span>
      </button>

      <p className="signin-card__alt">
        <Link href="/login" style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}><ArrowLeft size={14} strokeWidth={2} />{T('auth.backToLogin')}</Link>
      </p>
    </form>
  );
}
