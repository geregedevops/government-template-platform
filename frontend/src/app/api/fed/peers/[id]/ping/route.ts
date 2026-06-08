import { authedFetch } from '@/lib/api';
import { checkOrigin, proxyResult } from '@/lib/bff';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

/** POST /api/fed/peers/:id/ping — peer руу гарын үсэгтэй ping. */
export async function POST(req: Request, { params }: { params: { id: string } }) {
  const bad = checkOrigin(req);
  if (bad) return bad;
  const r = await authedFetch<unknown>(`/fed/peers/${encodeURIComponent(params.id)}/ping`, { method: 'POST' });
  return proxyResult(r);
}
