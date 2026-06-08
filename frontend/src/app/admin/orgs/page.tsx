import React from 'react';
import { redirect } from 'next/navigation';
import AppShell from '@/components/AppShell';
import OrgTree from '@/components/admin/OrgTree';
import { fetchMe, fetchMyPermissions } from '@/lib/api';
import { hasSession } from '@/lib/session';
import { initialsOf } from '@/lib/format';
import { t } from '@/lib/i18n';
import { getServerLang } from '@/lib/lang';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Байгууллага — Gerege' };

export default async function AdminOrgsPage() {
  if (!hasSession()) redirect('/login?next=/admin/orgs');
  const me = await fetchMe();
  if (!me) redirect('/login?next=/admin/orgs');
  const perms = await fetchMyPermissions();
  if (!perms.includes('org.manage')) redirect('/');

  const lang = getServerLang();

  return (
    <AppShell user={{ username: me.username, email: me.email, roleId: me.roleId, initials: initialsOf(me.username) }}>
      <div className="page-head">
        <span className="page-head__eyebrow">{t(lang, 'shell.sysAdmin')}</span>
        <h1>{t(lang, 'org.title')}</h1>
        <p className="page-head__sub">{t(lang, 'org.lede')}</p>
      </div>
      <OrgTree />
    </AppShell>
  );
}
