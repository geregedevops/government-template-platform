import React from 'react';
import { redirect } from 'next/navigation';
import AppShell from '@/components/AppShell';
import ProfileSections from '@/components/ProfileSections';
import { fetchMe, fetchMyPermissions } from '@/lib/api';
import { hasSession } from '@/lib/session';
import { initialsOf } from '@/lib/format';
import { t } from '@/lib/i18n';
import { getServerLang } from '@/lib/lang';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Менежер профайл — Gerege' };

export default async function ManagerProfilePage() {
  if (!hasSession()) redirect('/login?next=/manager/profile');
  const me = await fetchMe();
  if (!me) redirect('/login?next=/manager/profile');
  const perms = await fetchMyPermissions();
  if (!perms.includes('manager.view')) redirect('/');

  const lang = getServerLang();

  return (
    <AppShell user={{ username: me.username, email: me.email, roleId: me.roleId, initials: initialsOf(me.username) }}>
      <div className="page-head">
        <span className="page-head__eyebrow">{t(lang, 'shell.sysManager')}</span>
        <h1>{t(lang, 'nav.managerProfile')}</h1>
        <p className="page-head__sub">{t(lang, 'personal.profileLede')}</p>
      </div>
      <ProfileSections me={me} basePath="/manager" />
    </AppShell>
  );
}
