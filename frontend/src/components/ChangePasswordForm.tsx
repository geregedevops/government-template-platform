"use client";

import React, { useState } from 'react';
import { KeyRound } from 'lucide-react';
import Alert from '@/components/Alert';
import PasswordField from '@/components/PasswordField';
import { postJSON } from '@/lib/client';
import { useT } from '@/lib/useT';

export default function ChangePasswordForm() {
  const { T } = useT();
  const [current, setCurrent] = useState('');
  const [next, setNext] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError('');
    setSuccess(false);
    setFieldErrors({});

    const res = await postJSON('/api/auth/change-password', {
      current_password: current,
      new_password: next,
    });
    setBusy(false);

    if (res.ok) {
      setSuccess(true);
      setCurrent('');
      setNext('');
      return;
    }
    if (res.status === 422 && res.fieldErrors) {
      setFieldErrors(res.fieldErrors);
      return;
    }
    setError(res.message ?? T('auth.changePw.failed'));
  };

  return (
    <form className="form-grid" onSubmit={submit} noValidate>
      {success && <Alert kind="success">{T('auth.changePw.success')}</Alert>}
      {error && <Alert kind="danger">{error}</Alert>}

      <PasswordField
        label={T('auth.changePw.current')}
        value={current}
        onChange={setCurrent}
        autoComplete="current-password"
        error={fieldErrors.current_password}
        placeholder={T('auth.changePw.current')}
        name="current_password"
      />

      <PasswordField
        label={T('auth.reset.newPassword')}
        value={next}
        onChange={setNext}
        showStrength
        autoComplete="new-password"
        error={fieldErrors.new_password}
        placeholder={T('auth.reset.newPasswordPlaceholder')}
        name="new_password"
      />

      <div className="form-actions">
        <button className="btn btn--primary" type="submit" disabled={busy}>
          <KeyRound size={16} strokeWidth={2} />
          <span>{busy ? T('auth.changePw.submitting') : T('auth.changePw.submit')}</span>
        </button>
      </div>
    </form>
  );
}
