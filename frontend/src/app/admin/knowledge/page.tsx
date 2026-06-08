import { redirect } from 'next/navigation';
import AppShell from '@/components/AppShell';
import { hasSession } from '@/lib/session';
import { fetchMe, fetchMyPermissions } from '@/lib/api';
import { initialsOf } from '@/lib/format';
import { t } from '@/lib/i18n';
import { getServerLang } from '@/lib/lang';
import KnowledgeManager from '@/components/ai/KnowledgeManager';

export const dynamic = 'force-dynamic';

export const metadata = { title: 'Мэдлэг — Gerege' };

export default async function KnowledgePage() {
  if (!hasSession()) redirect('/login?next=/admin/knowledge');
  const me = await fetchMe();
  if (!me) redirect('/login?next=/admin/knowledge');
  // 'knowledge.manage' эрхтэй хэрэглэгч л нэвтэрнэ.
  const perms = await fetchMyPermissions();
  if (!perms.includes('knowledge.manage')) redirect('/');

  const lang = getServerLang();

  return (
    <AppShell user={{ username: me.username, email: me.email, roleId: me.roleId, initials: initialsOf(me.username) }}>
      <div className="page-head">
        <span className="page-head__eyebrow">{t(lang, 'kb.eyebrow')}</span>
        <h1>{t(lang, 'kb.title')}</h1>
        <p className="page-head__sub">{t(lang, 'kb.lede')}</p>
      </div>
      <KnowledgeManager />
    </AppShell>
  );
}
