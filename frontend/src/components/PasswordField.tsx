"use client";

import React, { useId } from 'react';
import { Check, X } from 'lucide-react';
import { evaluatePassword } from '@/lib/password';
import { useT } from '@/lib/useT';
import type { DictKey } from '@/lib/i18n';

interface Props {
  label: string;
  value: string;
  onChange: (v: string) => void;
  /** Хүчний хэмжигч + шаардлагын жагсаалт харуулах эсэх (бүртгэл/reset дээр). */
  showStrength?: boolean;
  error?: string;
  autoComplete?: string;
  placeholder?: string;
  name?: string;
}

const REQS: { key: keyof ReturnType<typeof evaluatePassword>['checks']; labelKey: DictKey }[] = [
  { key: 'length',  labelKey: 'pw.req.length' },
  { key: 'upper',   labelKey: 'pw.req.upper' },
  { key: 'lower',   labelKey: 'pw.req.lower' },
  { key: 'number',  labelKey: 'pw.req.number' },
  { key: 'special', labelKey: 'pw.req.special' },
];

const LEVEL_KEY: Record<'weak' | 'fair' | 'strong', DictKey> = {
  weak: 'pw.level.weak',
  fair: 'pw.level.fair',
  strong: 'pw.level.strong',
};

/**
 * Нууц үгийн талбар — нэмэлтээр хүчний хэмжигч + шаардлагын жагсаалттай.
 * Хэмжигчийн логик src/lib/password.ts дотор (тэндээс чангыг тохируулна).
 */
export default function PasswordField({
  label, value, onChange, showStrength, error, autoComplete = 'current-password',
  placeholder, name = 'password',
}: Props) {
  const { T } = useT();
  const id = useId();
  const strength = evaluatePassword(value);
  const segOn = strength.level === 'strong' ? 3 : strength.level === 'fair' ? 2 : value ? 1 : 0;

  return (
    <div className="field">
      <label className="field__label" htmlFor={id}>{label}</label>
      <input
        id={id}
        name={name}
        type="password"
        className="input"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        autoComplete={autoComplete}
        placeholder={placeholder}
        aria-invalid={error ? true : undefined}
        aria-describedby={showStrength ? `${id}-reqs` : undefined}
      />
      {error && <span className="field__error">{error}</span>}

      {showStrength && (
        <div className="pwmeter" id={`${id}-reqs`}>
          <div className="pwmeter__track" aria-hidden="true">
            {[0, 1, 2].map((i) => (
              <span
                key={i}
                className={`pwmeter__seg${i < segOn ? ` is-on--${strength.level}` : ''}`}
              />
            ))}
          </div>
          {value && (
            <span className="pwmeter__label">{T('pw.strength')}: {T(LEVEL_KEY[strength.level])}</span>
          )}
          <div className="pwmeter__reqs">
            {REQS.map((r) => {
              const met = strength.checks[r.key];
              return (
                <span key={r.key} className={`pwmeter__req${met ? ' is-met' : ''}`}>
                  {met ? <Check size={12} strokeWidth={3} /> : <X size={12} strokeWidth={2.5} />}
                  {T(r.labelKey)}
                </span>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}
