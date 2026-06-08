import { NextResponse } from 'next/server';
import { backendBaseURL, refreshAccessToken } from '@/lib/api';
import { getAccessToken } from '@/lib/session';
import { readJson, checkOrigin } from '@/lib/bff';
import { LANG_COOKIE } from '@/lib/i18n';
import { cookies } from 'next/headers';

export const dynamic = 'force-dynamic';
export const runtime = 'nodejs';

const TIMEOUT_MS = 30_000;

/** POST /api/voice/transcribe — дуу→бичвэр (STT) backend прокси. */
export async function POST(req: Request) {
  const bad = checkOrigin(req);
  if (bad) return bad;

  const body = await readJson<{ lang?: string; mime_type?: string; audio_base64?: string }>(req);
  if (!body.audio_base64 || !body.lang || !body.mime_type) {
    return NextResponse.json({ ok: false, status: 400, message: 'Аудио шаардлагатай.' }, { status: 400 });
  }

  const lang = cookies().get(LANG_COOKIE)?.value === 'en' ? 'en' : 'mn';
  const payload = JSON.stringify({ lang: body.lang, mime_type: body.mime_type, audio_base64: body.audio_base64 });

  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), TIMEOUT_MS);
  const call = (token: string | undefined) =>
    fetch(backendBaseURL() + '/voice/transcribe', {
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
      const t = await refreshAccessToken();
      if (t) res = await call(t);
    }
    type Env = { status?: boolean; message?: string; data?: unknown };
    let data: Env | null = null;
    try { data = (await res.json()) as Env; } catch {}
    if (!res.ok || !data || data.status === false) {
      const message = res.status >= 500 ? 'Дотоод алдаа гарлаа.' : data?.message ?? `Хүсэлт амжилтгүй (${res.status})`;
      return NextResponse.json({ ok: false, status: res.status, message }, { status: res.status >= 400 && res.status < 600 ? res.status : 502 });
    }
    return NextResponse.json({ ok: true, data: data.data }, { status: 200 });
  } catch (e: unknown) {
    const aborted = (e as { name?: string } | null)?.name === 'AbortError';
    return NextResponse.json(
      { ok: false, status: aborted ? 504 : 503, message: aborted ? 'Дуу таних хэт удаашрав.' : 'Backend-тэй холбогдож чадсангүй.' },
      { status: aborted ? 504 : 503 },
    );
  } finally {
    clearTimeout(timer);
  }
}
