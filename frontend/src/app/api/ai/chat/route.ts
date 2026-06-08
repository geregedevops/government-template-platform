import { NextResponse } from 'next/server';
import { backendBaseURL, refreshAccessToken } from '@/lib/api';
import { getAccessToken } from '@/lib/session';
import { readJson, checkOrigin } from '@/lib/bff';
import { LANG_COOKIE } from '@/lib/i18n';
import { cookies } from 'next/headers';

export const dynamic = 'force-dynamic';

// Streaming хариу нь Node runtime-д найдвартай (Edge-ийн fetch duplex
// хязгаарлалтаас зайлсхийнэ).
export const runtime = 'nodejs';

/** Streaming дуудлагын дээд хугацаа — backend-ийн AI timeout-оос арай урт. */
const STREAM_TIMEOUT_MS = 150_000;

/**
 * POST /api/ai/chat — backend /ai/chat (SSE) руу streaming прокси.
 * Токен httpOnly cookie-д байдаг тул browser backend руу шууд хандаж
 * чадахгүй — энэ route Bearer-ийг сервер талд хавсаргана. 401 ирвэл нэг
 * удаа refresh хийгээд дахин оролдоно (authedFetch-тэй ижил зарчим).
 */
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;

  const body = await readJson<{ conversation_id?: string; message?: string }>(req);
  if (!body.message || typeof body.message !== 'string' || body.message.trim() === '') {
    return NextResponse.json(
      { ok: false, status: 400, message: 'Мессеж шаардлагатай.' },
      { status: 400 },
    );
  }

  const lang = cookies().get(LANG_COOKIE)?.value === 'en' ? 'en' : 'mn';
  const payload = JSON.stringify({
    conversation_id: body.conversation_id || undefined,
    message: body.message,
  });

  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), STREAM_TIMEOUT_MS);

  const call = async (token: string | undefined) =>
    fetch(backendBaseURL() + '/ai/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'text/event-stream',
        // Backend-ийн i18n — алдааны мессеж хэрэглэгчийн хэлээр ирнэ.
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

    if (!res.ok || !res.body) {
      clearTimeout(timer);
      // Stream эхлээгүй алдаа — энгийн JSON болгон буцаана (клиент SSE
      // биш content-type-аар ялгана).
      let message = `Хүсэлт амжилтгүй (${res.status})`;
      try {
        const data = (await res.json()) as { message?: string };
        if (data?.message && res.status < 500) message = data.message;
      } catch {}
      return NextResponse.json(
        { ok: false, status: res.status, message },
        { status: res.status >= 400 && res.status < 600 ? res.status : 502 },
      );
    }

    // SSE body-г өөрчлөлгүй дамжуулна. Cleanup: stream дуусахад timer-ийг
    // цэвэрлэнэ (TransformStream flush).
    const cleanup = new TransformStream({
      flush() { clearTimeout(timer); },
    });
    return new Response(res.body.pipeThrough(cleanup), {
      status: 200,
      headers: {
        'Content-Type': 'text/event-stream; charset=utf-8',
        'Cache-Control': 'no-cache, no-transform',
        'X-Accel-Buffering': 'no',
      },
    });
  } catch (e: unknown) {
    clearTimeout(timer);
    const aborted = (e as { name?: string } | null)?.name === 'AbortError';
    return NextResponse.json(
      {
        ok: false,
        status: aborted ? 504 : 503,
        message: aborted ? 'AI хариу хэт удаашрав.' : 'Backend-тэй холбогдож чадсангүй.',
      },
      { status: aborted ? 504 : 503 },
    );
  }
}
