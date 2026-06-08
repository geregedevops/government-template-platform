import React from 'react';
import Link from 'next/link';
import AnonThemeToggle from './AnonThemeToggle';
import { t } from '@/lib/i18n';
import { getServerLang } from '@/lib/lang';

interface Props {
  /** topbar баруун талын нэмэлт навигаци (анхдагч: загвар солигч). */
  rightNav?: React.ReactNode;
  hideFooter?: boolean;
  children: React.ReactNode;
}

/**
 * Анонимос бүрхүүл — landing (/) болон auth хуудаснуудад. Rail / UserMenu / session
 * байхгүй. Брэнд topbar + төвлөрсөн агуулга + footer.
 */
export default function SigninShell({ rightNav, hideFooter, children }: Props) {
  const lang = getServerLang();
  return (
    <div className="signin-shell">
      <header className="signin-shell__nav">
        <Link className="topbar__brand" href="/">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img className="topbar__brand-mark" src="/brand.webp" alt="Gerege" />
          <div className="topbar__brand-text">
            <span className="topbar__brand-name">Gerege</span>
            <span className="topbar__brand-tag">{t(lang, 'shell.brandTag')}</span>
          </div>
        </Link>
        {rightNav ?? <AnonThemeToggle />}
      </header>

      <main className="signin-shell__body">{children}</main>

      {!hideFooter && (
        <footer className="signin-footer">
          <span>© 2026 Gerege Systems · <span className="mono">gerege-template-ai-v1.0</span></span>
          <span>
            <a href="https://gerege.mn/privacy" style={{ color: 'var(--muted)' }}>{t(lang, 'footer.privacy')}</a>
            <span style={{ padding: '0 8px', color: 'var(--faint)' }}>·</span>
            <a href="https://gerege.mn/terms" style={{ color: 'var(--muted)' }}>{t(lang, 'footer.terms')}</a>
          </span>
        </footer>
      )}
    </div>
  );
}
