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

export const metadata = { title: 'Хэрэглэгчийн профайл — Gerege' };

export default async function UserProfilePage() {
  if (!hasSession()) redirect('/login?next=/user/profile');
  const me = await fetchMe();
  if (!me) redirect('/login?next=/user/profile');
  const perms = await fetchMyPermissions();
  if (!perms.includes('personal.view')) redirect('/');

  const lang = getServerLang();

  return (
    <AppShell user={{ username: me.username, email: me.email, roleId: me.roleId, initials: initialsOf(me.username) }}>
      <div className="page-head">
        <span className="page-head__eyebrow">{t(lang, 'shell.sysUser')}</span>
        <h1>{t(lang, 'nav.personalProfile')}</h1>
        <p className="page-head__sub">{t(lang, 'personal.profileLede')}</p>
      </div>
      <ProfileSections me={me} basePath="/user" />
    </AppShell>
  );
}
