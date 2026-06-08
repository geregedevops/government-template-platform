"use client";

import React, { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { LogIn } from 'lucide-react';
import Alert from '@/components/Alert';
import PasswordField from '@/components/PasswordField';
import { postJSON } from '@/lib/client';
import { safeNext } from '@/lib/navigation';
import { useT } from '@/lib/useT';

export default function LoginForm({ next, notice }: { next: string; notice?: string }) {
  const router = useRouter();
  const { T } = useT();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [needsActivation, setNeedsActivation] = useState(false);

  const noticeText =
    notice === 'verified' ? T('login.notice.verified')
    : notice === 'registered' ? T('login.notice.registered')
    : notice === 'reset' ? T('login.notice.reset')
    : '';

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Browser-ийн autofill нь DOM талбарыг дүүргэдэг ч React-ийн onChange-г
    // үргэлж асаадаггүй тул controlled state хоосон үлдэж болзошгүй (талбарт
    // цэг харагдавч "талбар дутуу" гэж гарах шалтгаан). Тиймээс submit дээр
    // DOM-оос (FormData) бодит утгыг уншиж, state-ийг нөөц болгоно.
    const form = e.currentTarget as HTMLFormElement;
    const fd = new FormData(form);
    const emailVal = String(fd.get('email') ?? email).trim();
    const passwordVal = String(fd.get('password') ?? password);
    // Autofill-аас уншсан утгуудаар state-г синк хийнэ (доорх needsActivation
    // холбоос зэрэг UI зөв email ашиглана).
    if (emailVal !== email) setEmail(emailVal);
    if (passwordVal !== password) setPassword(passwordVal);

    setBusy(true);
    setError('');
    setFieldErrors({});
    setNeedsActivation(false);

    const res = await postJSON('/api/auth/login', { email: emailVal, password: passwordVal });
    setBusy(false);

    if (res.ok) {
      router.push(safeNext(next));
      router.refresh();
      return;
    }
    if (res.status === 422 && res.fieldErrors) {
      setFieldErrors(res.fieldErrors);
      return;
    }
    // Backend: идэвхжээгүй бүртгэл → 403 "account is not activated"
    // (mn хэл дээр "идэвхжээгүй" гэж орчуулагдана — хоёуланг нь таних).
    if (res.status === 403 && /(activat|идэвхж)/i.test(res.message ?? '')) {
      setNeedsActivation(true);
      return;
    }
    setError(res.message ?? T('login.failed'));
  };

  return (
    <form className="form-grid" onSubmit={submit} noValidate>
      {noticeText && <Alert kind="success">{noticeText}</Alert>}
      {error && <Alert kind="danger">{error}</Alert>}
      {needsActivation && (
        <Alert kind="info">
          {T('login.needsActivation')}{' '}
          <Link href={`/verify-otp?email=${encodeURIComponent(email)}`}>{T('login.verifyWithCode')}</Link>
        </Alert>
      )}

      <div className="field">
        <label className="field__label" htmlFor="email">{T('login.email')}</label>
        <input
          id="email"
          name="email"
          type="email"
          className="input"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          autoComplete="email"
          placeholder="tani@gerege.mn"
          aria-invalid={fieldErrors.email ? true : undefined}
          required
        />
        {fieldErrors.email && <span className="field__error">{fieldErrors.email}</span>}
      </div>

      <PasswordField
        label={T('login.password')}
        value={password}
        onChange={setPassword}
        autoComplete="current-password"
        error={fieldErrors.password}
        placeholder="••••••••••••"
      />

      <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: -6 }}>
        <Link href="/forgot-password" style={{ fontSize: 13 }}>{T('login.forgot')}</Link>
      </div>

      <button className="btn btn--primary btn--lg btn--block" type="submit" disabled={busy}>
        <LogIn size={18} strokeWidth={2} />
        <span>{busy ? T('login.submitting') : T('login.submit')}</span>
      </button>

      <p className="signin-card__alt">
        {T('login.noAccount')} <Link href="/register">{T('login.registerNow')}</Link>
      </p>
    </form>
  );
}
