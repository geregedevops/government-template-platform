import { authedFetch } from '@/lib/api';
import { checkOrigin, proxyResult } from '@/lib/bff';
import { LANG_COOKIE } from '@/lib/i18n';
import { cookies } from 'next/headers';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

interface Ctx {
  params: { id: string };
}

function lang(): string {
  return cookies().get(LANG_COOKIE)?.value === 'en' ? 'en' : 'mn';
}

/** DELETE /api/users/:id — хэрэглэгч устгах (admin). */
export async function DELETE(req: Request, { params }: Ctx) {
  const bad = checkOrigin(req);
  if (bad) return bad;
  const r = await authedFetch<unknown>(`/users/${encodeURIComponent(params.id)}`, {
    method: 'DELETE',
    headers: { 'Accept-Language': lang() },
  });
  return proxyResult(r);
}
