import { authedFetch } from '@/lib/api';
import { proxyResult } from '@/lib/bff';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

/** GET /api/fed/status — энэ node-ийн федерацийн төлөв. */
export async function GET() {
  const r = await authedFetch<unknown>('/fed/status', { method: 'GET' });
  return proxyResult(r);
}
