import { authedFetch } from '@/lib/api';
import { checkOrigin, proxyResult } from '@/lib/bff';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

/** POST /api/fed/flush — гарах дарааллыг нэн даруй боловсруулах. */
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;
  const r = await authedFetch<unknown>('/fed/flush', { method: 'POST' });
  return proxyResult(r);
}
