"use client";

import React, { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { UserPlus } from 'lucide-react';
import Alert from '@/components/Alert';
import PasswordField from '@/components/PasswordField';
import { postJSON } from '@/lib/client';
import { useT } from '@/lib/useT';

export default function RegisterForm() {
  const { T } = useT();
  const router = useRouter();
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError('');
    setFieldErrors({});

    const res = await postJSON('/api/auth/register', { username, email, password });

    if (res.ok) {
      // Бүртгэл идэвхгүй үүссэн — баталгаажуулалт руу шилжиж кодыг автоматаар илгээнэ.
      router.push(`/verify-otp?email=${encodeURIComponent(email)}&autosend=1`);
      return;
    }
    setBusy(false);
    if (res.status === 422 && res.fieldErrors) {
      setFieldErrors(res.fieldErrors);
      return;
    }
    // 409 = username/email давхцал.
    setError(res.message ?? T('auth.register.failed'));
  };

  return (
    <form className="form-grid" onSubmit={submit} noValidate>
      {error && <Alert kind="danger">{error}</Alert>}

      <div className="field">
        <label className="field__label" htmlFor="username">{T('auth.username')}</label>
        <input
          id="username"
          name="username"
          type="text"
          className="input"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          autoComplete="username"
          placeholder="johndoe"
          minLength={3}
          maxLength={25}
          aria-invalid={fieldErrors.username ? true : undefined}
          required
        />
        {fieldErrors.username
          ? <span className="field__error">{fieldErrors.username}</span>
          : <span className="field__help">{T('auth.username.help')}</span>}
      </div>

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
          maxLength={50}
          aria-invalid={fieldErrors.email ? true : undefined}
          required
        />
        {fieldErrors.email
          ? <span className="field__error">{fieldErrors.email}</span>
          : <span className="field__help">{T('auth.email.helpVerify')}</span>}
      </div>

      <PasswordField
        label={T('auth.password')}
        value={password}
        onChange={setPassword}
        showStrength
        autoComplete="new-password"
        error={fieldErrors.password}
        placeholder={T('auth.password.strong')}
      />

      <button className="btn btn--primary btn--lg btn--block" type="submit" disabled={busy}>
        <UserPlus size={18} strokeWidth={2} />
        <span>{busy ? T('auth.register.submitting') : T('auth.register.submit')}</span>
      </button>

      <p className="signin-card__alt">
        {T('auth.register.haveAccount')} <Link href="/login">{T('auth.signin')}</Link>
      </p>
    </form>
  );
}
