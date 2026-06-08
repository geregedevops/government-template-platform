import React from 'react';
import { redirect } from 'next/navigation';
import AppShell from '@/components/AppShell';
import ProfileSections from '@/components/ProfileSections';
import { fetchMe } from '@/lib/api';
import { initialsOf } from '@/lib/format';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Профайл — Gerege' };

export default async function ProfilePage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/admin/profile');

  const initials = initialsOf(me.username);

  return (
    <AppShell user={{ username: me.username, email: me.email, roleId: me.roleId, initials }}>
      <div className="page-head">
        <span className="page-head__eyebrow">Хувь хүн</span>
        <h1>Хувь хүний профайл</h1>
        <p className="page-head__sub">Нэвтэрсэн бүртгэлийн мэдээлэл. Эх сурвалж: <span className="mono">GET /users/me</span>.</p>
      </div>
      <ProfileSections me={me} />
    </AppShell>
  );
}
