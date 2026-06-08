import { authedFetch } from '@/lib/api';
import { readJson, checkOrigin, proxyResult } from '@/lib/bff';
import { LANG_COOKIE } from '@/lib/i18n';
import { cookies } from 'next/headers';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

function lang(): string {
  return cookies().get(LANG_COOKIE)?.value === 'en' ? 'en' : 'mn';
}

interface Ctx { params: { id: string }; }

/** PUT /api/orgs/:id — байгууллагын нэр/төрөл засах. */
export async function PUT(req: Request, { params }: Ctx) {
  const bad = checkOrigin(req);
  if (bad) return bad;
  const body = await readJson<Record<string, unknown>>(req);
  const r = await authedFetch<unknown>(`/organizations/${encodeURIComponent(params.id)}`, {
    method: 'PUT',
    headers: { 'Accept-Language': lang() },
    body: JSON.stringify(body),
  });
  return proxyResult(r);
}

/** DELETE /api/orgs/:id — байгууллага устгах (дэд мод бүхэлдээ). */
export async function DELETE(req: Request, { params }: Ctx) {
  const bad = checkOrigin(req);
  if (bad) return bad;
  const r = await authedFetch<unknown>(`/organizations/${encodeURIComponent(params.id)}`, {
    method: 'DELETE',
    headers: { 'Accept-Language': lang() },
  });
  return proxyResult(r);
}
