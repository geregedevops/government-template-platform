import { NextResponse } from 'next/server';
import { backendBaseURL, refreshAccessToken } from '@/lib/api';
import { getAccessToken } from '@/lib/session';
import { readJson, checkOrigin } from '@/lib/bff';
import { LANG_COOKIE } from '@/lib/i18n';
import { cookies } from 'next/headers';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

/**
 * Дуу хоолойн орчуулга нь Gemini руу 2 дуудлага (STT/орчуулга + TTS) хийдэг
 * тул backend-ийн 30с global timeout-оос арай урт BFF timeout тавина.
 */
const VOICE_TIMEOUT_MS = 45_000;

/**
 * POST /api/voice/translate — backend /voice/translate руу прокси. Аудио
 * base64-аар ирж, орчуулга + base64 WAV буцна. Токен httpOnly cookie-д
 * байдаг тул энэ route Bearer-ийг сервер талд хавсаргана; 401 ирвэл нэг удаа
 * refresh хийгээд дахин оролдоно (chat route-тэй ижил зарчим).
 */
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;

  const body = await readJson<{ source_lang?: string; mime_type?: string; audio_base64?: string }>(req);
  if (!body.audio_base64 || !body.source_lang || !body.mime_type) {
    return NextResponse.json(
      { ok: false, status: 400, message: 'Аудио болон эх хэл шаардлагатай.' },
      { status: 400 },
    );
  }

  const lang = cookies().get(LANG_COOKIE)?.value === 'en' ? 'en' : 'mn';
  const payload = JSON.stringify({
    source_lang: body.source_lang,
    mime_type: body.mime_type,
    audio_base64: body.audio_base64,
  });

  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), VOICE_TIMEOUT_MS);

  const call = async (token: string | undefined) =>
    fetch(backendBaseURL() + '/voice/translate', {
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

    type BackendEnvelope = { status?: boolean; message?: string; data?: unknown };
    let data: BackendEnvelope | null = null;
    try {
      data = (await res.json()) as BackendEnvelope;
    } catch {}

    if (!res.ok || !data || data.status === false) {
      // 5xx-ийн дотоод мессежийг гадагш гаргахгүй (info leak).
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
        message: aborted ? 'Орчуулга хэт удаашрав.' : 'Backend-тэй холбогдож чадсангүй.',
      },
      { status: aborted ? 504 : 503 },
    );
  } finally {
    clearTimeout(timer);
  }
}
