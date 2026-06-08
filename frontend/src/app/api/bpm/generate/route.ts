import { NextResponse } from 'next/server';
import { backendBaseURL, refreshAccessToken } from '@/lib/api';
import { getAccessToken } from '@/lib/session';
import { readJson, checkOrigin } from '@/lib/bff';
import { LANG_COOKIE } from '@/lib/i18n';
import { cookies } from 'next/headers';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

// AI generate нь Claude руу нэг дуудлага хийдэг тул JSON route-уудаас урт
// timeout (backend-ийн AI timeout-аас арай урт).
const GEN_TIMEOUT_MS = 60_000;

/**
 * POST /api/bpm/generate — текст тайлбараас AI-аар процесс үүсгэх. Backend нь
 * процессыг хадгалаад буцаана; client нь modeler руу шилжинэ. Токен httpOnly
 * cookie-д байдаг тул сервер талд Bearer хавсаргаж, 401 дээр нэг refresh хийнэ.
 */
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;

  const body = await readJson<{ description?: string }>(req);
  if (!body.description || !body.description.trim()) {
    return NextResponse.json(
      { ok: false, status: 400, message: 'Тайлбар шаардлагатай.' },
      { status: 400 },
    );
  }

  const lang = cookies().get(LANG_COOKIE)?.value === 'en' ? 'en' : 'mn';
  const payload = JSON.stringify({ description: body.description });

  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), GEN_TIMEOUT_MS);

  const call = (token: string | undefined) =>
    fetch(backendBaseURL() + '/bpm/generate', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
        'Accept-Language': lang,
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: payload,
      cache: 'no-store',
      signal: controller.signal,
    });

  try {
    let res = await call(getAccessToken());
    if (res.status === 401) {
      const newToken = await refreshAccessToken();
      if (newToken) res = await call(newToken);
    }

    type Envelope = { status?: boolean; message?: string; data?: unknown };
    let data: Envelope | null = null;
    try {
      data = (await res.json()) as Envelope;
    } catch {
      /* JSON биш */
    }

    if (!res.ok || !data || data.status === false) {
      const message =
        res.status >= 500
          ? 'Дотоод алдаа гарлаа. Дахин оролдоно уу.'
          : data?.message ?? `Хүсэлт амжилтгүй (${res.status})`;
      return NextResponse.json(
        { ok: false, status: res.status, message },
        { status: res.status >= 400 && res.status < 600 ? res.status : 502 },
      );
    }
    return NextResponse.json({ ok: true, data: data.data }, { status: 200 });
  } catch (e: unknown) {
    const aborted = (e as { name?: string } | null)?.name === 'AbortError';
    return NextResponse.json(
      {
        ok: false,
        status: aborted ? 504 : 503,
        message: aborted ? 'AI хариу хэт удаашрав.' : 'Backend-тэй холбогдож чадсангүй.',
      },
      { status: aborted ? 504 : 503 },
    );
  } finally {
    clearTimeout(timer);
  }
}
