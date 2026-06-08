import 'server-only';
import { cookies } from 'next/headers';
import { LANG_COOKIE, type LangPref } from './i18n';

/**
 * Server component / route handler-т хэрэглэгчийн хэлийг уншина.
 * setLang (preferences.ts) нь localStorage-тэй зэрэгцээ httpOnly БИШ
 * cookie бичдэг тул server render хэрэглэгчийн сонгосон хэлээр гарна.
 * Cookie байхгүй (анхны зочлолт) бол 'mn' — платформын үндсэн хэл.
 */
export function getServerLang(): LangPref {
  const v = cookies().get(LANG_COOKIE)?.value;
  return v === 'en' ? 'en' : 'mn';
}
