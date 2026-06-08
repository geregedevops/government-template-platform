import { redirect } from 'next/navigation';
import AppShell from '@/components/AppShell';
import { hasSession } from '@/lib/session';
import { fetchMe, authedFetch } from '@/lib/api';
import { initialsOf } from '@/lib/format';
import { emptyBpmn, type BpmProcess } from '@/lib/bpm';
import ModelerCanvas from '@/components/bpm/ModelerCanvas';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Процесс зураглал — Gerege' };

interface Props {
  params: { id: string };
}

export default async function ModelerPage({ params }: Props) {
  const { id } = params;
  if (!hasSession()) redirect(`/login?next=/admin/bpm/modeler/${id}`);

  const me = await fetchMe();
  if (!me) redirect(`/login?next=/admin/bpm/modeler/${id}`);

  // Шинэ процесс бол хоосон BPMN диаграмм; засах бол backend-ээс татна.
  let initial: {
    id?: string;
    name: string;
    description: string;
    bpmn: string;
  } = { name: '', description: '', bpmn: emptyBpmn() };

  if (id !== 'new') {
    const r = await authedFetch<BpmProcess>(`/bpm/processes/${id}`, { method: 'GET' });
    if (!r.ok || !r.data) redirect('/admin/bpm');
    initial = {
      id: r.data.id,
      name: r.data.name,
      description: r.data.description,
      bpmn: r.data.bpmn || emptyBpmn(),
    };
  }

  return (
    <AppShell user={{ username: me.username, email: me.email, roleId: me.roleId, initials: initialsOf(me.username) }}>
      <ModelerCanvas initial={initial} />
    </AppShell>
  );
}
