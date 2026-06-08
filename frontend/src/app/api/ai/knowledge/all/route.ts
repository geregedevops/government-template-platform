import { authedFetch } from '@/lib/api';
import { proxyResult } from '@/lib/bff';
import { LANG_COOKIE } from '@/lib/i18n';
import { cookies } from 'next/headers';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

/** GET /api/ai/knowledge/all — бүх системийн мэдлэг (admin RLS). */
export async function GET() {
  const lang = cookies().get(LANG_COOKIE)?.value === 'en' ? 'en' : 'mn';
  const r = await authedFetch<unknown>('/ai/knowledge/all', {
    method: 'GET',
    headers: { 'Accept-Language': lang },
  });
  return proxyResult(r);
}
