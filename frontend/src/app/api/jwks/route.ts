export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

// GET /api/jwks (rewritten from /.well-known/jwks.json) — backend-ийн үндэс дэх
// JWKS-ийг прокси хийнэ. Backend JWKS нь /api/v1 дор биш үндэст байгаа тул
// authedFetch (которое /api/v1 нэмдэг)-ийн оронд шууд fetch ашиглана. Public,
// нэвтрэлтгүй (нийтийн түлхүүр) — токен дамжуулахгүй.
export async function GET() {
  const base = (process.env.BACKEND_URL ?? 'http://localhost:8080').replace(/\/$/, '');
  try {
    const r = await fetch(`${base}/.well-known/jwks.json`, { cache: 'no-store' });
    const body = await r.text();
    return new Response(body, {
      status: r.status,
      headers: { 'Content-Type': 'application/json', 'Cache-Control': 'public, max-age=300' },
    });
  } catch {
    return new Response('{"keys":[]}', { status: 200, headers: { 'Content-Type': 'application/json' } });
  }
}
