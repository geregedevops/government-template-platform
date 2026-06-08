import { authedFetch } from '@/lib/api';
import { readJson, checkOrigin, proxyResult } from '@/lib/bff';
import { LANG_COOKIE } from '@/lib/i18n';
import { cookies } from 'next/headers';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

interface Ctx {
  params: { id: string };
}

/** POST /api/bpm/tasks/:id/submit — form даалгаврыг бөглөж дараагийн алхам руу. */
export async function POST(req: Request, { params }: Ctx) {
  const bad = checkOrigin(req);
  if (bad) return bad;

  const lang = cookies().get(LANG_COOKIE)?.value === 'en' ? 'en' : 'mn';
  const body = await readJson<Record<string, unknown>>(req);
  const r = await authedFetch<unknown>(`/bpm/tasks/${encodeURIComponent(params.id)}/submit`, {
    method: 'POST',
    headers: { 'Accept-Language': lang },
    body: JSON.stringify(body),
  });
  return proxyResult(r);
}
