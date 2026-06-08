import { backendFetch } from '@/lib/api';
import { setSession } from '@/lib/session';
import { readJson, toClientResponse, checkOrigin, requireFields } from '@/lib/bff';
import type { BackendUser } from '@/lib/types';

export const dynamic = 'force-dynamic';

// POST /api/auth/login — backend /auth/login руу прокси, амжилттай бол токен
// хосыг httpOnly cookie-д суулгана. Токен хэзээ ч browser руу буцахгүй.
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;

  const body = await readJson<{ email?: string; password?: string }>(req);
  const missing = requireFields(body, ['email', 'password']);
  if (missing) return missing;

  // Клиентийн IP-г backend руу дамжуулна (lockout-ийг email+IP-ээр түлхүүрлэх).
  // ЗӨВХӨН ИТГЭМЖТЭЙ эх сурвалжийг авна: nginx нь `X-Real-IP $remote_addr`-ийг
  // ДАХИЖ БИЧДЭГ (overwrite) тул клиент хуурамчлах боломжгүй. X-Forwarded-For-ийн
  // ЭХНИЙ утга нь клиентийн илгээсэн (хуурамч) утга байж болзошгүй тул түүнийг
  // АВАХГҮЙ — авбал халдагч IP сэлгэн lockout-ийг тойрно. XFF-ийн СҮҮЛИЙН хоп
  // (nginx нэмсэн бодит IP)-ийг л fallback болгоно.
  const xff = (req.headers.get('x-forwarded-for') ?? '').split(',').map((s) => s.trim()).filter(Boolean);
  const clientIp = (req.headers.get('x-real-ip') ?? '').trim() || (xff.length ? xff[xff.length - 1] : '');

  const result = await backendFetch<BackendUser>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email: body.email, password: body.password }),
    headers: clientIp ? { 'X-Client-IP': clientIp } : undefined,
  });

  if (result.ok && result.data?.token && result.data?.refresh_token) {
    setSession(result.data.token, result.data.refresh_token);
  }

  return toClientResponse(result);
}
