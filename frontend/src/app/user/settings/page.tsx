import React from 'react';
import { redirect } from 'next/navigation';
import AppShell from '@/components/AppShell';
import SettingsSections from '@/components/SettingsSections';
import { fetchMe, fetchMyPermissions } from '@/lib/api';
import { hasSession } from '@/lib/session';
import { initialsOf } from '@/lib/format';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Аюулгүй байдал — Gerege' };

export default async function UserSettingsPage() {
  if (!hasSession()) redirect('/login?next=/user/settings');
  const me = await fetchMe();
  if (!me) redirect('/login?next=/user/settings');
  const perms = await fetchMyPermissions();
  if (!perms.includes('personal.view')) redirect('/');

  return (
    <AppShell user={{ username: me.username, email: me.email, roleId: me.roleId, initials: initialsOf(me.username) }}>
      <div className="page-head">
        <span className="page-head__eyebrow">Аюулгүй байдал</span>
        <h1>Аюулгүй байдлын тохиргоо</h1>
        <p className="page-head__sub">Нууц үгээ солих, идэвхтэй сессиэ удирдах.</p>
      </div>
      <SettingsSections />
    </AppShell>
  );
}
