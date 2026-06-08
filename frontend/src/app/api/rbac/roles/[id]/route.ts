import { authedFetch } from '@/lib/api';
import { readJson, checkOrigin, proxyResult } from '@/lib/bff';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

interface Ctx { params: { id: string } }

/** PUT /api/rbac/roles/:id — эрх засах. */
export async function PUT(req: Request, { params }: Ctx) {
  const bad = checkOrigin(req);
  if (bad) return bad;
  const body = await readJson<Record<string, unknown>>(req);
  const r = await authedFetch<unknown>(`/rbac/roles/${encodeURIComponent(params.id)}`, { method: 'PUT', body: JSON.stringify(body) });
  return proxyResult(r);
}

/** DELETE /api/rbac/roles/:id — эрх устгах. */
export async function DELETE(req: Request, { params }: Ctx) {
  const bad = checkOrigin(req);
  if (bad) return bad;
  const r = await authedFetch<unknown>(`/rbac/roles/${encodeURIComponent(params.id)}`, { method: 'DELETE' });
  return proxyResult(r);
}
