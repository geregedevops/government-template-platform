"use client";

import React from 'react';
import { LogOut } from 'lucide-react';
import { signOut } from '@/lib/signout';
import { showToast } from '@/lib/preferences';
import { useT } from '@/lib/useT';

/** Энэ төхөөрөмж дээрх сессийг хаах товч (server component-д ашиглах client wrapper). */
export default function SignOutButton({ label }: { label?: string }) {
  const { T } = useT();
  const handle = async () => {
    // Амжилтгүй (backend 5xx) бол чимээгүй өнгөрөхгүй — toast харуулна.
    const ok = await signOut();
    if (!ok) showToast(T('auth.signOutError'));
  };
  return (
    <button className="btn btn--danger" type="button" onClick={() => void handle()}>
      <LogOut size={16} strokeWidth={2} />
      <span>{label ?? T('auth.signOutDevice')}</span>
    </button>
  );
}
