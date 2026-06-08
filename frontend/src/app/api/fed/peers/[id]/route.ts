import { authedFetch } from '@/lib/api';
import { readJson, checkOrigin, proxyResult } from '@/lib/bff';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

/** PUT /api/fed/peers/:id — peer засах. */
export async function PUT(req: Request, { params }: { params: { id: string } }) {
  const bad = checkOrigin(req);
  if (bad) return bad;
  const body = await readJson<Record<string, unknown>>(req);
  const r = await authedFetch<unknown>(`/fed/peers/${encodeURIComponent(params.id)}`, {
    method: 'PUT',
    body: JSON.stringify(body),
  });
  return proxyResult(r);
}

/** DELETE /api/fed/peers/:id — peer устгах. */
export async function DELETE(req: Request, { params }: { params: { id: string } }) {
  const bad = checkOrigin(req);
  if (bad) return bad;
  const r = await authedFetch<unknown>(`/fed/peers/${encodeURIComponent(params.id)}`, { method: 'DELETE' });
  return proxyResult(r);
}
