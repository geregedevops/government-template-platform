import { authedFetch } from '@/lib/api';
import { proxyResult } from '@/lib/bff';
import { LANG_COOKIE } from '@/lib/i18n';
import { cookies } from 'next/headers';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

interface Ctx {
  params: { id: string };
}

/**
 * GET /api/bpm/instances/:id/task — гүйлтийн идэвхтэй даалгаврыг (рендерлэх
 * дэлгэц) авах. GET тул checkOrigin тавихгүй (Origin толгой ирэхгүй).
 */
export async function GET(_req: Request, { params }: Ctx) {
  const lang = cookies().get(LANG_COOKIE)?.value === 'en' ? 'en' : 'mn';
  const r = await authedFetch<unknown>(`/bpm/instances/${encodeURIComponent(params.id)}/task`, {
    method: 'GET',
    headers: { 'Accept-Language': lang },
  });
  return proxyResult(r);
}
