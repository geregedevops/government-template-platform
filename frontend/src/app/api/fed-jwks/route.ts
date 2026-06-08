export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

// GET /api/fed-jwks (rewritten from /.well-known/fed-jwks.json) — node-ийн
// гарын үсгийн нийтийн түлхүүрийг backend-ийн үндэс дэх fed-jwks-ээс прокси.
export async function GET() {
  const base = (process.env.BACKEND_URL ?? 'http://localhost:8080').replace(/\/$/, '');
  try {
    const r = await fetch(`${base}/.well-known/fed-jwks.json`, { cache: 'no-store' });
    const text = await r.text();
    return new Response(text, { status: r.status, headers: { 'Content-Type': 'application/json', 'Cache-Control': 'public, max-age=300' } });
  } catch {
    return new Response('{"keys":[]}', { status: 200, headers: { 'Content-Type': 'application/json' } });
  }
}
