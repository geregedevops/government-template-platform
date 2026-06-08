import { authedFetch } from '@/lib/api';
import { readJson, checkOrigin, proxyResult } from '@/lib/bff';
import { LANG_COOKIE } from '@/lib/i18n';
import { cookies } from 'next/headers';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

interface Ctx { params: { id: string }; }

function lang(): string {
  return cookies().get(LANG_COOKIE)?.value === 'en' ? 'en' : 'mn';
}

/** PATCH /api/users/:id/org — хэрэглэгчийг байгууллагад шилжүүлэх (admin). */
export async function PATCH(req: Request, { params }: Ctx) {
  const bad = checkOrigin(req);
  if (bad) return bad;
  const body = await readJson<Record<string, unknown>>(req);
  const r = await authedFetch<unknown>(`/users/${encodeURIComponent(params.id)}/org`, {
    method: 'PATCH',
    headers: { 'Accept-Language': lang() },
    body: JSON.stringify(body),
  });
  return proxyResult(r);
}
