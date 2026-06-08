import React from 'react';
import Link from 'next/link';
import { User, Mail, ShieldCheck, Clock, Hash, RefreshCw } from 'lucide-react';
import { roleLabel } from '@/lib/types';
import type { SessionUser } from '@/lib/types';
import { formatTS, initialsOf } from '@/lib/format';
import { getServerLang } from '@/lib/lang';
import { t } from '@/lib/i18n';

/**
 * Хэрэглэгчийн профайлын картууд — /profile ба /personal/profile хоёулаа
 * хуваалцана (нэг эх сурвалж: GET /users/me).
 */
export default function ProfileSections({ me, basePath = '/admin' }: { me: SessionUser; basePath?: string }) {
  const initials = initialsOf(me.username);
  const lang = getServerLang();
  return (
    <>
      <section className="card" aria-label={t(lang, 'profile.overview')}>
        <div className="profile-card">
          <div className="profile-card__avatar" aria-hidden="true">{initials}</div>
          <div className="profile-card__body">
            <div className="profile-card__name">
              <span className="profile-card__name-text">{me.username}</span>
              <span className="badge badge--primary">{roleLabel(me.roleId)}</span>
            </div>
            <div className="profile-card__sub">
              <span className="mono">{me.email}</span>
            </div>
          </div>
          <div className="profile-card__action">
            <Link className="btn btn--secondary" href={`${basePath}/settings`}>{t(lang, 'profile.changePw')}</Link>
          </div>
        </div>
      </section>

      <section className="card" aria-label={t(lang, 'profile.accountFields')}>
        <div className="card__head card__head--with-sub">
          <div className="card__title"><h2>{t(lang, 'profile.accountFields')}</h2></div>
          <span className="card__sub">{t(lang, 'profile.backendManaged')}</span>
        </div>

        <div>
          <div className="defrow">
            <span className="defrow__label"><Hash size={13} style={{ verticalAlign: 'middle', marginRight: 6 }} />ID</span>
            <span className="defrow__value mono">{me.id}</span>
          </div>
          <div className="defrow">
            <span className="defrow__label"><User size={13} style={{ verticalAlign: 'middle', marginRight: 6 }} />{t(lang, 'profile.username')}</span>
            <span className="defrow__value">{me.username}</span>
          </div>
          <div className="defrow">
            <span className="defrow__label"><Mail size={13} style={{ verticalAlign: 'middle', marginRight: 6 }} />{t(lang, 'profile.email')}</span>
            <span className="defrow__value mono">{me.email}</span>
          </div>
          <div className="defrow">
            <span className="defrow__label"><ShieldCheck size={13} style={{ verticalAlign: 'middle', marginRight: 6 }} />{t(lang, 'profile.role')}</span>
            <span className="defrow__value">
              <span className="chip chip--neutral">role_id {me.roleId}</span> {roleLabel(me.roleId)}
            </span>
          </div>
          <div className="defrow">
            <span className="defrow__label"><Clock size={13} style={{ verticalAlign: 'middle', marginRight: 6 }} />{t(lang, 'profile.createdAt')}</span>
            <span className="defrow__value mono">{formatTS(me.createdAt)}</span>
          </div>
          <div className="defrow">
            <span className="defrow__label"><RefreshCw size={13} style={{ verticalAlign: 'middle', marginRight: 6 }} />{t(lang, 'profile.updatedAt')}</span>
            <span className="defrow__value mono">{me.updatedAt ? formatTS(me.updatedAt) : '—'}</span>
          </div>
        </div>
      </section>
    </>
  );
}
