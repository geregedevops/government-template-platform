import { redirect } from 'next/navigation';
import { User, ShieldCheck } from 'lucide-react';
import AppShell from '@/components/AppShell';
import DashboardView, { type DashboardCard } from '@/components/DashboardView';
import { hasSession } from '@/lib/session';
import { fetchMe, fetchMyPermissions } from '@/lib/api';
import { initialsOf } from '@/lib/format';
import { getServerLang } from '@/lib/lang';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Хэрэглэгчийн самбар — Gerege' };

const CARDS: DashboardCard[] = [
  { href: '/user/profile',  eyebrowKey: 'card.profile.eyebrow',  titleKey: 'card.profile.title',  descKey: 'card.profile.desc',  icon: User },
  { href: '/user/settings', eyebrowKey: 'card.security.eyebrow', titleKey: 'card.security.title', descKey: 'card.security.desc', icon: ShieldCheck },
];

export default async function UserDashboardPage() {
  if (!hasSession()) redirect('/login?next=/user/dashboard');
  const me = await fetchMe();
  if (!me) redirect('/login?next=/user/dashboard');
  const perms = await fetchMyPermissions();
  if (!perms.includes('personal.view')) redirect('/');
  const lang = getServerLang();

  return (
    <AppShell user={{ username: me.username, email: me.email, roleId: me.roleId, initials: initialsOf(me.username) }}>
      <DashboardView me={me} lang={lang} profileHref="/user/profile" cards={CARDS} />
    </AppShell>
  );
}
