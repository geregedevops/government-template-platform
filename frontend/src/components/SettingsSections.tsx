import React from 'react';
import { KeyRound, LogOut } from 'lucide-react';
import ChangePasswordForm from '@/components/ChangePasswordForm';
import SignOutButton from '@/components/SignOutButton';
import { getServerLang } from '@/lib/lang';
import { t } from '@/lib/i18n';

/**
 * Аюулгүй байдлын тохиргооны картууд (нууц үг солих + сесси) — Admin / Manager /
 * Personal систем бүрийн settings хуудас хуваалцана.
 */
export default function SettingsSections() {
  const lang = getServerLang();
  return (
    <>
      <section className="card" aria-label={t(lang, 'settings.changePwTitle')}>
        <div className="card__head card__head--with-sub">
          <div className="card__title">
            <KeyRound size={18} strokeWidth={2} style={{ color: 'var(--dan-blue-text)' }} />
            <h2>{t(lang, 'settings.changePwTitle')}</h2>
          </div>
          <span className="card__sub">{t(lang, 'settings.changePwDesc')}</span>
        </div>
        <ChangePasswordForm />
      </section>

      <section className="card" aria-label={t(lang, 'settings.sessionTitle')}>
        <div className="card__head card__head--with-sub">
          <div className="card__title">
            <LogOut size={18} strokeWidth={2} style={{ color: 'var(--danger)' }} />
            <h2>{t(lang, 'settings.sessionTitle')}</h2>
          </div>
          <span className="card__sub">{t(lang, 'settings.sessionDesc')}</span>
        </div>
        <SignOutButton />
      </section>
    </>
  );
}
