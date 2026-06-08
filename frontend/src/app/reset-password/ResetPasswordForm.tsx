"use client";

import React, { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { KeyRound, ArrowLeft } from 'lucide-react';
import Alert from '@/components/Alert';
import PasswordField from '@/components/PasswordField';
import { postJSON } from '@/lib/client';
import { useT } from '@/lib/useT';

export default function ResetPasswordForm({ initialToken }: { initialToken: string }) {
  const { T } = useT();
  const router = useRouter();
  const [token, setToken] = useState(initialToken);
  const [password, setPassword] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError('');
    setFieldErrors({});
    const res = await postJSON('/api/auth/reset-password', { token, new_password: password });
    setBusy(false);
    if (res.ok) {
      router.push('/login?notice=reset');
      return;
    }
    if (res.status === 422 && res.fieldErrors) {
      setFieldErrors(res.fieldErrors);
      return;
    }
    setError(res.message ?? T('auth.reset.failed'));
  };

  return (
    <form className="form-grid" onSubmit={submit} noValidate>
      {error && <Alert kind="danger">{error}</Alert>}

      {!initialToken && (
        <div className="field">
          <label className="field__label" htmlFor="token">{T('auth.reset.token')}</label>
          <input
            id="token"
            name="token"
            className="input mono"
            value={token}
            onChange={(e) => setToken(e.target.value)}
            placeholder={T('auth.reset.tokenPlaceholder')}
            aria-invalid={fieldErrors.token ? true : undefined}
            required
          />
          {fieldErrors.token && <span className="field__error">{fieldErrors.token}</span>}
        </div>
      )}

      <PasswordField
        label={T('auth.reset.newPassword')}
        value={password}
        onChange={setPassword}
        showStrength
        autoComplete="new-password"
        error={fieldErrors.new_password}
        placeholder={T('auth.reset.newPasswordPlaceholder')}
      />

      <button className="btn btn--primary btn--lg btn--block" type="submit" disabled={busy || !token}>
        <KeyRound size={18} strokeWidth={2} />
        <span>{busy ? T('auth.reset.submitting') : T('auth.reset.submit')}</span>
      </button>

      <p className="signin-card__alt">
        <Link href="/login" style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}><ArrowLeft size={14} strokeWidth={2} />{T('auth.backToLogin')}</Link>
      </p>
    </form>
  );
}
