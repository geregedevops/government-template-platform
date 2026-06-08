import { redirect } from 'next/navigation';
import AppShell from '@/components/AppShell';
import { hasSession } from '@/lib/session';
import { fetchMe } from '@/lib/api';
import { initialsOf } from '@/lib/format';
import { t } from '@/lib/i18n';
import { getServerLang } from '@/lib/lang';
import RunClient from '@/components/bpm/RunClient';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Процесс ажиллуулах — Gerege' };

interface Props {
  params: { id: string };
}

export default async function RunPage({ params }: Props) {
  const { id } = params;
  if (!hasSession()) redirect(`/login?next=/admin/bpm/run/${id}`);

  const me = await fetchMe();
  if (!me) redirect(`/login?next=/admin/bpm/run/${id}`);

  const lang = getServerLang();

  return (
    <AppShell user={{ username: me.username, email: me.email, roleId: me.roleId, initials: initialsOf(me.username) }}>
      <div className="page-head">
        <span className="page-head__eyebrow">{t(lang, 'bpm.run.eyebrow')}</span>
        <h1>{t(lang, 'bpm.title')}</h1>
        <p className="page-head__sub">{t(lang, 'bpm.run.lede')}</p>
      </div>
      <RunClient processId={id} />
    </AppShell>
  );
}
