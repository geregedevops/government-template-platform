import { NextResponse } from 'next/server';
import { authedFetch } from '@/lib/api';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

/**
 * GET /api/voice/history — backend /voice/history руу прокси. Хурдан JSON
 * дуудлага тул authedFetch-ийн стандарт timeout/refresh-ийг ашиглана.
 *
 * Энд checkOrigin тавихгүй: browser нь same-origin GET fetch дээр Origin
 * толгой илгээдэггүй тул checkOrigin нь хууль ёсны дуудлагыг 403 хийнэ.
 * Хамгаалалт нь SameSite=strict session cookie дээр тулгуурлана — cross-site
 * хүсэлт токен авч явахгүй тул authedFetch 401 болж, хувийн түүх алдагдахгүй.
 */
export async function GET() {
  const r = await authedFetch<unknown>('/voice/history', { method: 'GET' });
  if (!r.ok) {
    const message = r.status >= 500 ? 'Дотоод алдаа гарлаа.' : r.message;
    return NextResponse.json({ ok: false, status: r.status, message }, { status: r.status });
  }
  return NextResponse.json({ ok: true, data: r.data ?? [] }, { status: 200 });
}
