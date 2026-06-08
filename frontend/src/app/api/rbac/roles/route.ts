import { authedFetch } from '@/lib/api';
import { readJson, checkOrigin, proxyResult } from '@/lib/bff';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

/** GET /api/rbac/roles — эрхүүд (permission-уудтай). */
export async function GET() {
  const r = await authedFetch<unknown>('/rbac/roles', { method: 'GET' });
  return proxyResult(r);
}

/** POST /api/rbac/roles — эрх үүсгэх. */
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;
  const body = await readJson<Record<string, unknown>>(req);
  const r = await authedFetch<unknown>('/rbac/roles', { method: 'POST', body: JSON.stringify(body) });
  return proxyResult(r);
}
