import React from 'react';
import { redirect } from 'next/navigation';
import AppShell from '@/components/AppShell';
import { hasSession } from '@/lib/session';
import { fetchMe, fetchMyPermissions } from '@/lib/api';
import { initialsOf } from '@/lib/format';
import { t } from '@/lib/i18n';
import { getServerLang } from '@/lib/lang';
import ChatClient from './ChatClient';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'AI чат — Gerege' };

export default async function ChatPage() {
  if (!hasSession()) redirect('/login?next=/admin/chat');

  const me = await fetchMe();
  if (!me) redirect('/login?next=/admin/chat');
  // 'ai.chat' эрхтэй хэрэглэгч л нэвтэрнэ (backend ч мөн адил хамгаалалттай).
  const perms = await fetchMyPermissions();
  if (!perms.includes('ai.chat')) redirect('/');

  const lang = getServerLang();

  return (
    <AppShell user={{ username: me.username, email: me.email, roleId: me.roleId, initials: initialsOf(me.username) }}>
      <div className="page-head">
        <span className="page-head__eyebrow">{t(lang, 'chat.eyebrow')}</span>
        <h1>{t(lang, 'chat.title')}</h1>
        <p className="page-head__sub">{t(lang, 'chat.lede')}</p>
      </div>
      <ChatClient />
    </AppShell>
  );
}
