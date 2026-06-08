import React from 'react';
import Link from 'next/link';
import { ChevronRight, User, ShieldCheck, KeyRound, Info, Mail, Clock, type LucideIcon } from 'lucide-react';
import { roleLabel } from '@/lib/types';
import type { SessionUser } from '@/lib/types';
import { formatDate, formatTS, formatWeekday, initialsOf } from '@/lib/format';
import { t, type DictKey, type LangPref } from '@/lib/i18n';

export interface DashboardCard {
  href: string;
  eyebrowKey: DictKey;
  titleKey: DictKey;
  descKey: DictKey;
  icon: LucideIcon;
}

interface Props {
  me: SessionUser;
  lang: LangPref;
  profileHref: string;
  cards: DashboardCard[];
}

/**
 * Admin / Manager / User системийн самбарын нэгдсэн дүр төрх — мэндчилгээ,
 * профайлын тойм, хэсгийн картууд, бүртгэлийн дэлгэрэнгүй. Систем бүр өөрийн
 * profileHref + cards-ыг дамжуулна.
 */
export default function DashboardView({ me, lang, profileHref, cards }: Props) {
  const today = new Date();
  const initials = initialsOf(me.username);

  return (
    <>
      <div className="page-head">
        <span className="page-head__eyebrow">{t(lang, 'home.eyebrow')}</span>
        <h1>{t(lang, 'home.greeting')} {me.username}</h1>
        <p className="page-head__sub">
          {formatDate(lang, today)}, {formatWeekday(lang, today)} · <span className="mono">UTC+08</span>
        </p>
      </div>

      <section className="card" aria-label={t(lang, 'home.profileOverview')}>
        <div className="profile-card">
          <div className="profile-card__avatar" aria-hidden="true">{initials}</div>
          <div className="profile-card__body">
            <div className="profile-card__name">
              <span className="profile-card__name-text">{me.username}</span>
              <span className="badge badge--primary">{roleLabel(me.roleId, lang)}</span>
            </div>
            <div className="profile-card__sub">
              <span className="mono">{me.email}</span>
              <span className="dot" />
              <span>{t(lang, 'home.registeredAt')} <span className="mono">{formatTS(me.createdAt)}</span></span>
            </div>
          </div>
          <div className="profile-card__action">
            <Link className="btn btn--secondary" href={profileHref}>{t(lang, 'home.viewProfile')}</Link>
          </div>
        </div>
      </section>

      <div className="section-divider">{t(lang, 'home.sections')}</div>

      <div className="grid-2">
        {cards.map((c) => {
          const Icon = c.icon;
          return (
            <Link
              key={c.href}
              href={c.href}
              className="card"
              style={{ textDecoration: 'none', color: 'inherit', display: 'flex', flexDirection: 'column', gap: 10 }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <div style={{ width: 36, height: 36, borderRadius: 10, display: 'grid', placeItems: 'center', background: 'var(--dan-blue-soft)', color: 'var(--dan-blue-text)' }}>
                  <Icon size={18} strokeWidth={2} />
                </div>
                <span className="page-head__eyebrow">{t(lang, c.eyebrowKey)}</span>
              </div>
              <h3 style={{ fontSize: 16, fontWeight: 600 }}>{t(lang, c.titleKey)}</h3>
              <p style={{ fontSize: 13, color: 'var(--muted)', lineHeight: 1.55 }}>{t(lang, c.descKey)}</p>
              <span style={{ marginTop: 'auto', display: 'inline-flex', alignItems: 'center', gap: 4, fontSize: 12, fontWeight: 500, color: 'var(--dan-blue-text)' }}>
                {t(lang, 'home.open')} <ChevronRight size={12} strokeWidth={2} />
              </span>
            </Link>
          );
        })}
      </div>

      <section className="card" aria-label={t(lang, 'home.accountDetails')} style={{ marginTop: 16 }}>
        <div className="card__head card__head--with-sub">
          <div className="card__title">
            <Info size={18} strokeWidth={2} style={{ color: 'var(--dan-blue-text)' }} />
            <h2>{t(lang, 'home.accountDetails')}</h2>
          </div>
          <span className="card__sub">gerege-template-ai-v1.0 · GET /users/me</span>
        </div>
        <div>
          <div className="defrow">
            <span className="defrow__label"><User size={13} style={{ verticalAlign: 'middle', marginRight: 6 }} />{t(lang, 'home.username')}</span>
            <span className="defrow__value">{me.username}</span>
          </div>
          <div className="defrow">
            <span className="defrow__label"><Mail size={13} style={{ verticalAlign: 'middle', marginRight: 6 }} />{t(lang, 'home.email')}</span>
            <span className="defrow__value mono">{me.email}</span>
          </div>
          <div className="defrow">
            <span className="defrow__label"><ShieldCheck size={13} style={{ verticalAlign: 'middle', marginRight: 6 }} />{t(lang, 'home.role')}</span>
            <span className="defrow__value"><span className="chip chip--neutral">role_id {me.roleId}</span> {roleLabel(me.roleId, lang)}</span>
          </div>
          <div className="defrow">
            <span className="defrow__label"><Clock size={13} style={{ verticalAlign: 'middle', marginRight: 6 }} />{t(lang, 'home.createdAt')}</span>
            <span className="defrow__value mono">{formatTS(me.createdAt)}</span>
          </div>
        </div>
      </section>

      <div className="trust-strip">
        <span className="trust-strip__item">
          <KeyRound size={12} strokeWidth={2.5} style={{ color: 'var(--dan-blue)' }} />
          JWT access + refresh
        </span>
        <span className="trust-strip__dot">·</span>
        <span className="trust-strip__item">bcrypt</span>
        <span className="trust-strip__dot">·</span>
        <span className="trust-strip__item">Fiber v3 + GORM</span>
        <span className="trust-strip__dot">·</span>
        <span className="trust-strip__item mono">TLS 1.3</span>
      </div>

      <footer className="footer">
        <span>© 2026 Gerege Systems · <span className="mono">gerege-template-ai-v1.0</span></span>
        <span className="footer__links">
          <a href="https://gerege.mn/privacy">{t(lang, 'footer.privacy')}</a>
          <a href="https://gerege.mn/terms">{t(lang, 'footer.terms')}</a>
        </span>
      </footer>
    </>
  );
}
