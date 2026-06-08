import { authedFetch } from '@/lib/api';
import { proxyResult } from '@/lib/bff';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

/** GET /api/rbac/permissions — эрхийн каталог (admin). */
export async function GET() {
  const r = await authedFetch<unknown>('/rbac/permissions', { method: 'GET' });
  return proxyResult(r);
}
