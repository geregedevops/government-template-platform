"use client";

import { useCallback, useEffect, useSyncExternalStore } from 'react';

export type ThemePref = 'light' | 'dark' | 'system';
export type LangPref = 'mn' | 'en';

const KEYS = { theme: 'gerege.theme', lang: 'gerege.lang' } as const;
const VALID = {
  theme: new Set<ThemePref>(['light', 'dark', 'system']),
  lang: new Set<LangPref>(['mn', 'en']),
};

const read = <T extends string>(key: 'theme' | 'lang', fallback: T, valid: Set<T>): T => {
  if (typeof window === 'undefined') return fallback;
  try {
    const v = localStorage.getItem(KEYS[key]) as T | null;
    return v && valid.has(v) ? v : fallback;
  } catch {
    return fallback;
  }
};

const applyTheme = (value: ThemePref) => {
  if (typeof document === 'undefined') return;
  const effective: 'light' | 'dark' =
    value === 'system'
      ? window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
      : value;
  const root = document.documentElement;
  if (effective === 'dark') root.setAttribute('data-theme', 'dark');
  else root.removeAttribute('data-theme');
  root.setAttribute('data-theme-pref', value);
};

const applyLang = (value: LangPref) => {
  if (typeof document === 'undefined') return;
  document.documentElement.setAttribute('lang', value);
  // Server component-ууд (SSR render) хэлийг мэдэхийн тулд localStorage-тэй
  // зэрэгцээ энгийн (httpOnly биш) cookie-д давхар бичнэ — src/lib/lang.ts-ийн
  // getServerLang() үүнийг уншина. Нууц мэдээлэл биш тул JS-д ил байх нь OK.
  try {
    document.cookie = `gerege.lang=${value}; path=/; max-age=31536000; samesite=lax`;
  } catch {}
};

// --- Хуваалцсан store (бүх usePreferences хэрэглэгч нэг эх сурвалжаас уншина) ---
// Өмнө нь component бүр өөрийн useState-тэй байсан тул нэг газар хэл/загвар
// солиход бусад нь хуучнаар үлддэг байв. useSyncExternalStore-оор нэг
// модуль-түвшний төлөвт холбож, өөрчлөлт БҮХ хэрэглэгчид тархана.
let _theme: ThemePref = 'light';
let _lang: LangPref = 'mn';
let _hydrated = false;
const listeners = new Set<() => void>();
const emit = () => { for (const l of listeners) l(); };
const subscribe = (cb: () => void) => {
  listeners.add(cb);
  return () => { listeners.delete(cb); };
};

/**
 * gerege-ийн тохиргоог (загвар + хэл) localStorage-д уншиж/бичээд <html> дээр
 * тусгана. Бүх компонент хуваалцсан store-оос уншина — нэг газар солиход
 * бүгд шинэчлэгдэнэ.
 */
export function usePreferences() {
  // SSR + эхний client render нь анхдагч (mn/light)-аар — hydration зөрөхгүй;
  // дараа нь эхний mount localStorage-аас синк хийж emit хийнэ.
  const theme = useSyncExternalStore(subscribe, () => _theme, () => 'light' as ThemePref);
  const lang = useSyncExternalStore(subscribe, () => _lang, () => 'mn' as LangPref);

  useEffect(() => {
    if (_hydrated) return;
    _hydrated = true;
    _theme = read('theme', 'light', VALID.theme);
    _lang = read<LangPref>('lang', 'mn', VALID.lang);
    applyTheme(_theme);
    applyLang(_lang);
    emit();
  }, []);

  // OS загвар солигдоход "system" дээр байвал дахин тусгана.
  useEffect(() => {
    if (theme !== 'system' || typeof window === 'undefined') return;
    const mql = window.matchMedia('(prefers-color-scheme: dark)');
    const handler = () => applyTheme('system');
    mql.addEventListener('change', handler);
    return () => mql.removeEventListener('change', handler);
  }, [theme]);

  const setTheme = useCallback((value: ThemePref) => {
    if (!VALID.theme.has(value)) return;
    _theme = value;
    try { localStorage.setItem(KEYS.theme, value); } catch {}
    applyTheme(value);
    emit();
  }, []);

  const setLang = useCallback((value: LangPref) => {
    if (!VALID.lang.has(value)) return;
    _lang = value;
    try { localStorage.setItem(KEYS.lang, value); } catch {}
    applyLang(value);
    emit();
  }, []);

  return { theme, setTheme, lang, setLang };
}

/** Жижиг toast туслах — globals.css дахь .toast класс ашиглана. */
export function showToast(message: string) {
  if (typeof document === 'undefined') return;
  let el = document.querySelector<HTMLDivElement>('.toast[data-app-toast]');
  if (!el) {
    el = document.createElement('div');
    el.className = 'toast';
    el.dataset.appToast = '1';
    el.setAttribute('role', 'status');
    el.setAttribute('aria-live', 'polite');
    document.body.appendChild(el);
  }
  el.textContent = message;
  requestAnimationFrame(() => el!.classList.add('is-visible'));
  window.clearTimeout((el as unknown as { _t: number })._t);
  (el as unknown as { _t: number })._t = window.setTimeout(() => el!.classList.remove('is-visible'), 1800);
}
