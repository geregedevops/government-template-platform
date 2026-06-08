import { redirect } from 'next/navigation';
import AppShell from '@/components/AppShell';
import { hasSession } from '@/lib/session';
import { fetchMe, fetchMyPermissions } from '@/lib/api';
import { initialsOf } from '@/lib/format';
import { t } from '@/lib/i18n';
import { getServerLang } from '@/lib/lang';
import ProcessList from '@/components/bpm/ProcessList';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Процесс — Gerege' };

export default async function BpmPage() {
  if (!hasSession()) redirect('/login?next=/admin/bpm');

  const me = await fetchMe();
  if (!me) redirect('/login?next=/admin/bpm');
  // 'bpm.manage' эрхтэй хэрэглэгч л нэвтэрнэ (backend ч мөн адил хамгаалалттай).
  const perms = await fetchMyPermissions();
  if (!perms.includes('bpm.manage')) redirect('/');

  const lang = getServerLang();

  return (
    <AppShell user={{ username: me.username, email: me.email, roleId: me.roleId, initials: initialsOf(me.username) }}>
      <div className="page-head">
        <span className="page-head__eyebrow">{t(lang, 'bpm.eyebrow')}</span>
        <h1>{t(lang, 'bpm.title')}</h1>
        <p className="page-head__sub">{t(lang, 'bpm.lede')}</p>
      </div>
      <ProcessList />
    </AppShell>
  );
}
