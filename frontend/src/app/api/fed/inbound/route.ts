export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

// POST /api/fed/inbound — peer-ээс ирэх гарын үсэгтэй мессежийг backend руу
// дамжуулна. Нэвтрэлтгүй (machine-to-machine; итгэлийг гарын үсэг тогтооно).
export async function POST(req: Request) {
  const base = (process.env.BACKEND_URL ?? 'http://localhost:8080').replace(/\/$/, '');
  const body = await req.text();
  try {
    const r = await fetch(`${base}/api/v1/fed/inbound`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/jose' },
      body,
      cache: 'no-store',
    });
    const text = await r.text();
    return new Response(text, { status: r.status, headers: { 'Content-Type': 'application/json' } });
  } catch {
    return new Response('{"error":"upstream unavailable"}', { status: 502, headers: { 'Content-Type': 'application/json' } });
  }
}
