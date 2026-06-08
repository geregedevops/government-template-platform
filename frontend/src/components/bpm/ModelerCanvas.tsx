"use client";

import dynamic from 'next/dynamic';
import { Loader2 } from 'lucide-react';

// bpmn-js / form-js нь window-д хамаардаг тул ssr:false-ээр зөвхөн client дээр
// ачаална (SSR үед render хийхгүй).
const BpmModeler = dynamic(() => import('./BpmModeler'), {
  ssr: false,
  loading: () => (
    <div className="bpm-run__state muted">
      <Loader2 size={18} strokeWidth={2} className="spin" />
    </div>
  ),
});

interface Props {
  initial: {
    id?: string;
    name: string;
    description: string;
    bpmn: string;
  };
}

export default function ModelerCanvas({ initial }: Props) {
  return <BpmModeler initial={initial} />;
}
