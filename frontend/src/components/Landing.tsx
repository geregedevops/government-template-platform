import React from 'react';
import Link from 'next/link';
import { LogIn, KeyRound, Info } from 'lucide-react';
import SigninShell from '@/components/SigninShell';
import { t, type LangPref } from '@/lib/i18n';

/** Нийтийн landing — нэвтрээгүй зочдод харагдах нүүр (root зам дээр). */
export default function Landing({ lang }: { lang: LangPref }) {
  return (
    <SigninShell>
      <section className="signin-card" aria-labelledby="landing-title">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img className="signin-card__crest" src="/brand.webp" alt="" aria-hidden="true" />

        <div>
          <div className="page-head__eyebrow" style={{ marginBottom: 6 }}>{t(lang, 'landing.eyebrow')}</div>
          <h1 id="landing-title">Gerege</h1>
          <p className="signin-card__lede" style={{ marginTop: 12 }}>
            {t(lang, 'landing.lede')}
          </p>
        </div>

        <Link className="btn btn--primary btn--lg btn--block" href="/login" aria-label={t(lang, 'landing.signInAria')}>
          <LogIn size={18} strokeWidth={2} />
          <span>{t(lang, 'landing.signIn')}</span>
        </Link>

        <p className="signin-card__alt">
          {t(lang, 'landing.noAccount')} <Link href="/register">{t(lang, 'landing.registerNow')}</Link>
        </p>

        <p className="signin-card__helper">
          <Info size={14} strokeWidth={2} />
          <span>{t(lang, 'landing.helper')}</span>
        </p>

        <div className="signin-card__trust" aria-label={t(lang, 'landing.trustLabel')}>
          <span className="badge"><KeyRound size={11} strokeWidth={2} /> JWT</span>
          <span className="badge">bcrypt</span>
          <span className="badge">Fiber v3 + GORM</span>
          <span className="badge"><span className="mono" style={{ fontSize: 11 }}>TLS 1.3</span></span>
        </div>
      </section>
    </SigninShell>
  );
}
