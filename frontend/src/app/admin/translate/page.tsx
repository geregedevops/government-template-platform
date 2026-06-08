import React from 'react';
import { redirect } from 'next/navigation';
import AppShell from '@/components/AppShell';
import { hasSession } from '@/lib/session';
import { fetchMe, fetchMyPermissions } from '@/lib/api';
import { initialsOf } from '@/lib/format';
import { t } from '@/lib/i18n';
import { getServerLang } from '@/lib/lang';
import TranslateClient from './TranslateClient';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Дуу хоолойн орчуулга — Gerege' };

export default async function TranslatePage() {
  if (!hasSession()) redirect('/login?next=/admin/translate');

  const me = await fetchMe();
  if (!me) redirect('/login?next=/admin/translate');
  // 'voice.translate' эрхтэй хэрэглэгч л нэвтэрнэ.
  const perms = await fetchMyPermissions();
  if (!perms.includes('voice.translate')) redirect('/');

  const lang = getServerLang();

  return (
    <AppShell user={{ username: me.username, email: me.email, roleId: me.roleId, initials: initialsOf(me.username) }}>
      <div className="page-head">
        <span className="page-head__eyebrow">{t(lang, 'translate.eyebrow')}</span>
        <h1>{t(lang, 'translate.title')}</h1>
        <p className="page-head__sub">{t(lang, 'translate.lede')}</p>
      </div>
      <TranslateClient />
    </AppShell>
  );
}
