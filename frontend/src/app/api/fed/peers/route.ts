import { authedFetch } from '@/lib/api';
import { readJson, checkOrigin, proxyResult } from '@/lib/bff';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

/** GET /api/fed/peers — гишүүн node-уудыг жагсаах. */
export async function GET() {
  const r = await authedFetch<unknown>('/fed/peers', { method: 'GET' });
  return proxyResult(r);
}

/** POST /api/fed/peers — гишүүн node бүртгэх. */
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;
  const body = await readJson<Record<string, unknown>>(req);
  const r = await authedFetch<unknown>('/fed/peers', { method: 'POST', body: JSON.stringify(body) });
  return proxyResult(r);
}
