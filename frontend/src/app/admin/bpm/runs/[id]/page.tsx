import { redirect } from 'next/navigation';
import Link from 'next/link';
import { ArrowLeft } from 'lucide-react';
import AppShell from '@/components/AppShell';
import { hasSession } from '@/lib/session';
import { fetchMe, authedFetch } from '@/lib/api';
import { initialsOf } from '@/lib/format';
import { t } from '@/lib/i18n';
import { getServerLang } from '@/lib/lang';
import { type BpmProcess } from '@/lib/bpm';
import InstanceList from '@/components/bpm/InstanceList';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Гүйлтүүд — Gerege' };

interface Props {
  params: { id: string };
}

export default async function RunsPage({ params }: Props) {
  const { id } = params;
  if (!hasSession()) redirect(`/login?next=/admin/bpm/runs/${id}`);

  const me = await fetchMe();
  if (!me) redirect(`/login?next=/admin/bpm/runs/${id}`);

  const lang = getServerLang();
  // Процессыг авч эзэмшил/нэрийг шалгана (өөр хэрэглэгчийнх бол /bpm руу).
  const r = await authedFetch<BpmProcess>(`/bpm/processes/${id}`, { method: 'GET' });
  if (!r.ok || !r.data) redirect('/admin/bpm');

  return (
    <AppShell user={{ username: me.username, email: me.email, roleId: me.roleId, initials: initialsOf(me.username) }}>
      <div className="page-head">
        <span className="page-head__eyebrow">{t(lang, 'bpm.runs.eyebrow')}</span>
        <h1>{r.data.name}</h1>
        <p className="page-head__sub">{t(lang, 'bpm.runs.lede')}</p>
      </div>
      <div style={{ marginBottom: 16 }}>
        <Link className="btn btn--ghost" href="/admin/bpm">
          <ArrowLeft size={16} strokeWidth={2} />
          <span>{t(lang, 'bpm.modeler.back')}</span>
        </Link>
      </div>
      <InstanceList processId={id} />
    </AppShell>
  );
}
