import { redirect } from 'next/navigation';
import { User, ShieldCheck, Sparkles } from 'lucide-react';
import AppShell from '@/components/AppShell';
import DashboardView, { type DashboardCard } from '@/components/DashboardView';
import { hasSession } from '@/lib/session';
import { fetchMe } from '@/lib/api';
import { initialsOf } from '@/lib/format';
import { getServerLang } from '@/lib/lang';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Хяналтын самбар — Gerege' };

const CARDS: DashboardCard[] = [
  { href: '/admin/profile',  eyebrowKey: 'card.profile.eyebrow',  titleKey: 'card.profile.title',  descKey: 'card.profile.desc',  icon: User },
  { href: '/admin/chat',     eyebrowKey: 'card.ai.eyebrow',       titleKey: 'card.ai.title',       descKey: 'card.ai.desc',       icon: Sparkles },
  { href: '/admin/settings', eyebrowKey: 'card.security.eyebrow', titleKey: 'card.security.title', descKey: 'card.security.desc', icon: ShieldCheck },
];

export default async function AdminDashboardPage() {
  if (!hasSession()) redirect('/login?next=/admin/dashboard');
  const me = await fetchMe();
  if (!me) redirect('/login?next=/admin/dashboard');
  const lang = getServerLang();

  return (
    <AppShell user={{ username: me.username, email: me.email, roleId: me.roleId, initials: initialsOf(me.username) }}>
      <DashboardView me={me} lang={lang} profileHref="/admin/profile" cards={CARDS} />
    </AppShell>
  );
}
