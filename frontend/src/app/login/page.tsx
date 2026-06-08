import React from 'react';
import SigninShell from '@/components/SigninShell';
import { safeNext } from '@/lib/navigation';
import { t } from '@/lib/i18n';
import { getServerLang } from '@/lib/lang';
import LoginForm from './LoginForm';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Нэвтрэх — Gerege' };

export default function LoginPage({
  searchParams,
}: {
  searchParams: { next?: string; notice?: string };
}) {
  const lang = getServerLang();
  const next = safeNext(searchParams.next);

  return (
    <SigninShell>
      <section className="signin-card signin-card--narrow" aria-labelledby="login-title">
        <div>
          <div className="page-head__eyebrow" style={{ marginBottom: 6 }}>{t(lang, 'login.eyebrow')}</div>
          <h1 id="login-title">{t(lang, 'login.title')}</h1>
          <p className="signin-card__lede" style={{ marginTop: 8, fontSize: 14 }}>
            {t(lang, 'login.lede')}
          </p>
        </div>
        <LoginForm next={next} notice={searchParams.notice} />
      </section>
    </SigninShell>
  );
}
