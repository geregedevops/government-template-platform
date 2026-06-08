"use client";

import { useCallback } from 'react';
import { usePreferences } from './preferences';
import { t, type DictKey } from './i18n';

/**
 * Client component-уудын орчуулгын hook. usePreferences-ийн lang төлөвт
 * суурилдаг тул хэл солигдоход компонентууд шууд дахин render хийгдэнэ.
 *
 *   const { T, lang } = useT();
 *   <span>{T('nav.dashboard')}</span>
 */
export function useT() {
  const { lang, setLang } = usePreferences();
  const T = useCallback((key: DictKey) => t(lang, key), [lang]);
  return { T, lang, setLang };
}
