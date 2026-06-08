import { redirect } from 'next/navigation';
import { User, ShieldCheck } from 'lucide-react';
import AppShell from '@/components/AppShell';
import DashboardView, { type DashboardCard } from '@/components/DashboardView';
import { hasSession } from '@/lib/session';
import { fetchMe, fetchMyPermissions } from '@/lib/api';
import { initialsOf } from '@/lib/format';
import { getServerLang } from '@/lib/lang';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Менежер самбар — Gerege' };

const CARDS: DashboardCard[] = [
  { href: '/manager/profile',  eyebrowKey: 'card.profile.eyebrow',  titleKey: 'card.profile.title',  descKey: 'card.profile.desc',  icon: User },
  { href: '/manager/settings', eyebrowKey: 'card.security.eyebrow', titleKey: 'card.security.title', descKey: 'card.security.desc', icon: ShieldCheck },
];

export default async function ManagerDashboardPage() {
  if (!hasSession()) redirect('/login?next=/manager/dashboard');
  const me = await fetchMe();
  if (!me) redirect('/login?next=/manager/dashboard');
  const perms = await fetchMyPermissions();
  if (!perms.includes('manager.view')) redirect('/');
  const lang = getServerLang();

  return (
    <AppShell user={{ username: me.username, email: me.email, roleId: me.roleId, initials: initialsOf(me.username) }}>
      <DashboardView me={me} lang={lang} profileHref="/manager/profile" cards={CARDS} />
    </AppShell>
  );
}
