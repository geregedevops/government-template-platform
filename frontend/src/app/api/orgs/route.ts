import { authedFetch } from '@/lib/api';
import { readJson, checkOrigin, proxyResult } from '@/lib/bff';
import { LANG_COOKIE } from '@/lib/i18n';
import { cookies } from 'next/headers';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

function lang(): string {
  return cookies().get(LANG_COOKIE)?.value === 'en' ? 'en' : 'mn';
}

/** GET /api/orgs — байгууллагын модыг жагсаах. */
export async function GET() {
  const r = await authedFetch<unknown>('/organizations', { method: 'GET', headers: { 'Accept-Language': lang() } });
  return proxyResult(r);
}

/** POST /api/orgs — байгууллага үүсгэх (parent доор). */
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;
  const body = await readJson<Record<string, unknown>>(req);
  const r = await authedFetch<unknown>('/organizations', {
    method: 'POST',
    headers: { 'Accept-Language': lang() },
    body: JSON.stringify(body),
  });
  return proxyResult(r);
}
