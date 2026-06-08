import React from 'react';
import { redirect } from 'next/navigation';
import AppShell from '@/components/AppShell';
import SettingsSections from '@/components/SettingsSections';
import { fetchMe } from '@/lib/api';
import { hasSession } from '@/lib/session';
import { initialsOf } from '@/lib/format';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Аюулгүй байдал — Gerege' };

export default async function AdminSettingsPage() {
  if (!hasSession()) redirect('/login?next=/admin/settings');
  const me = await fetchMe();
  if (!me) redirect('/login?next=/admin/settings');

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
