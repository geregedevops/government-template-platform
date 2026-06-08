import { authedFetch } from '@/lib/api';
import { checkOrigin, proxyResult } from '@/lib/bff';
import { LANG_COOKIE } from '@/lib/i18n';
import { cookies } from 'next/headers';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

interface Ctx {
  params: { id: string };
}

/** POST /api/bpm/processes/:id/start — процессын шинэ гүйлт эхлүүлнэ. */
export async function POST(req: Request, { params }: Ctx) {
  const bad = checkOrigin(req);
  if (bad) return bad;

  const lang = cookies().get(LANG_COOKIE)?.value === 'en' ? 'en' : 'mn';
  const r = await authedFetch<unknown>(`/bpm/processes/${encodeURIComponent(params.id)}/start`, {
    method: 'POST',
    headers: { 'Accept-Language': lang },
  });
  return proxyResult(r);
}
