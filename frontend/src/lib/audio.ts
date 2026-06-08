"use client";

// Browser-ийн дуу бичлэгийн хуваалцсан туслахууд — чат (STT) болон
// /translate хуудас хоёул ашиглана.

// Backend-ийн VOICE_MAX_AUDIO_KB (640 KiB)-тэй тааруулсан client-side хязгаар:
// base64 + JSON нь глобал 1 MiB body cap дотор багтах ёстой.
export const MAX_AUDIO_BYTES = 620 * 1024;
// Авто-зогсолт: урт бичлэгээс (хэмжээ + timeout) сэргийлнэ.
export const MAX_RECORD_MS = 45_000;

/** Browser-ийн дэмждэг хамгийн тохиромжтой бичлэгийн форматыг сонгоно. */
export function pickMimeType(): string {
  if (typeof MediaRecorder === 'undefined') return '';
  const candidates = ['audio/webm;codecs=opus', 'audio/webm', 'audio/mp4', 'audio/ogg;codecs=opus'];
  for (const c of candidates) {
    if (MediaRecorder.isTypeSupported(c)) return c;
  }
  return '';
}

/** MIME-ийн codecs параметрийг хасч суурь төрлийг буцаана (backend allowlist). */
export function baseMime(mime: string): string {
  return mime.split(';')[0] || 'audio/webm';
}

// TTS-ийн дээд урт — урт текст Gemini TTS-д ажлын хугацааны төсвөөс (≈25с)
// хэтэрч timeout болдог тул чатын "Сонсох" дээр энэ хүртэл уншина.
export const TTS_MAX_CHARS = 700;

/**
 * Markdown тэмдэглэгээг цэвэр текст болгоно — TTS нь `**`, `##`, `-`, `` ` ``
 * зэргийг чанга унших ёсгүй. Мөн TTS_MAX_CHARS-аар богиносгоно (урт хариуг
 * эхнээс нь уншина; timeout-аас сэргийлнэ).
 */
export function stripMarkdown(md: string): string {
  let t = md
    .replace(/```[\s\S]*?```/g, ' ')        // код блок
    .replace(/`([^`]+)`/g, '$1')            // inline код
    .replace(/!\[[^\]]*\]\([^)]*\)/g, ' ')  // зураг
    .replace(/\[([^\]]+)\]\([^)]*\)/g, '$1') // холбоос → текст
    .replace(/^\s{0,3}#{1,6}\s+/gm, '')     // гарчиг #
    .replace(/^\s{0,3}>\s?/gm, '')          // quote
    .replace(/^\s*[-*+]\s+/gm, '')          // bullet
    .replace(/^\s*\d+\.\s+/gm, '')          // дугаарласан жагсаалт
    .replace(/(\*\*|__)(.*?)\1/g, '$2')     // bold
    .replace(/(\*|_)(.*?)\1/g, '$2')        // italic
    .replace(/\n{2,}/g, '. ')               // догол мөр → завсарлага
    .replace(/\s+/g, ' ')
    .trim();
  if (t.length > TTS_MAX_CHARS) t = t.slice(0, TTS_MAX_CHARS).trim() + '…';
  return t;
}

/** Blob-г base64 болгоно (data: угтварыг хасна). */
export function blobToBase64(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onloadend = () => resolve((reader.result as string).split(',')[1] ?? '');
    reader.onerror = reject;
    reader.readAsDataURL(blob);
  });
}
