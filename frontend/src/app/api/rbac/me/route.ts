import { authedFetch } from '@/lib/api';
import { proxyResult } from '@/lib/bff';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

/** GET /api/rbac/me — одоогийн хэрэглэгчийн эрхүүд (цэс шүүхэд). */
export async function GET() {
  const r = await authedFetch<unknown>('/rbac/me', { method: 'GET' });
  return proxyResult(r);
}
