import React from 'react';
import { redirect } from 'next/navigation';
import AppShell from '@/components/AppShell';
import UsersManager from '@/components/admin/UsersManager';
import { fetchMe, fetchMyPermissions } from '@/lib/api';
import { hasSession } from '@/lib/session';
import { initialsOf } from '@/lib/format';
import { t } from '@/lib/i18n';
import { getServerLang } from '@/lib/lang';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Хэрэглэгчид — Gerege' };

export default async function AdminUsersPage() {
  if (!hasSession()) redirect('/login?next=/admin/users');
  const me = await fetchMe();
  if (!me) redirect('/login?next=/admin/users');
  // 'users.manage' эрхтэй хэрэглэгч л нэвтэрнэ (backend ч мөн адил хамгаалалттай).
  const perms = await fetchMyPermissions();
  if (!perms.includes('users.manage')) redirect('/');

  const lang = getServerLang();

  return (
    <AppShell user={{ username: me.username, email: me.email, roleId: me.roleId, initials: initialsOf(me.username) }}>
      <div className="page-head">
        <span className="page-head__eyebrow">{t(lang, 'shell.sysAdmin')}</span>
        <h1>{t(lang, 'nav.users')}</h1>
        <p className="page-head__sub">{t(lang, 'users.lede')}</p>
      </div>
      <UsersManager currentUserId={me.id} />
    </AppShell>
  );
}
