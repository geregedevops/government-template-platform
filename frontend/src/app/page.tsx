import { redirect } from 'next/navigation';
import Landing from '@/components/Landing';
import { hasSession } from '@/lib/session';
import { fetchMe, fetchMyPermissions } from '@/lib/api';
import { getServerLang } from '@/lib/lang';

export const dynamic = 'force-dynamic';

/**
 * Үндсэн зам — зочдод нийтийн Landing; нэвтэрсэн хэрэглэгчийг эрхээр нь зөв
 * систем рүү шилжүүлнэ (admin → /admin/dashboard, менежер → /manager,
 * бусад → /personal).
 */
export default async function RootPage() {
  const lang = getServerLang();
  if (!hasSession()) return <Landing lang={lang} />;

  const me = await fetchMe();
  if (!me) return <Landing lang={lang} />;

  const perms = await fetchMyPermissions();
  if (perms.includes('dashboard.view')) redirect('/admin/dashboard');
  if (perms.includes('manager.view')) redirect('/manager/dashboard');
  redirect('/user/dashboard');
}
