"use client";

import React, { useEffect, useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  LayoutDashboard, User, ShieldCheck, HelpCircle, LogOut, Sparkles, Languages, Workflow,
  BookOpen, Menu, Search, Users, ShieldHalf, UserCircle, Briefcase, Building2,
} from 'lucide-react';
import UserMenu from './UserMenu';
import { signOut } from '@/lib/signout';
import { useT } from '@/lib/useT';
import type { DictKey } from '@/lib/i18n';

const ROLE_ADMIN = 1; // backend domain.RoleAdmin

export interface AppUser {
  username: string;
  email: string;
  initials: string;
  roleId: number;
}

interface Props {
  user: AppUser;
  children: React.ReactNode;
}

interface NavItem {
  href: string;
  labelKey: DictKey;
  icon: typeof User;
  perm?: string; // шаардагдах эрх; байхгүй бол бүх нэвтэрсэн хэрэглэгчид
}
interface NavGroup {
  labelKey?: DictKey;
  items: NavItem[];
}
// Систем = icon rail дахь дээд түвшний бүлэг. adminOnly бол зөвхөн admin харна.
interface NavSystem {
  key: string;
  labelKey: DictKey;
  icon: typeof User;
  adminOnly: boolean;
  groups: NavGroup[];
}

const SYSTEMS: NavSystem[] = [
  {
    key: 'admin',
    labelKey: 'shell.sysAdmin',
    icon: ShieldHalf,
    adminOnly: true,
    groups: [
      {
        labelKey: 'shell.groupProfile',
        items: [
          { href: '/admin/dashboard', labelKey: 'nav.dashboard', icon: LayoutDashboard, perm: 'dashboard.view' },
          { href: '/admin/profile', labelKey: 'nav.profile', icon: User, perm: 'dashboard.view' },
        ],
      },
      {
        labelKey: 'shell.groupAI',
        items: [
          { href: '/admin/chat', labelKey: 'nav.chat', icon: Sparkles, perm: 'ai.chat' },
          { href: '/admin/knowledge', labelKey: 'nav.knowledge', icon: BookOpen, perm: 'knowledge.manage' },
          { href: '/admin/translate', labelKey: 'nav.translate', icon: Languages, perm: 'voice.translate' },
          { href: '/admin/bpm', labelKey: 'nav.bpm', icon: Workflow, perm: 'bpm.manage' },
        ],
      },
      {
        labelKey: 'shell.groupSecurity',
        items: [
          { href: '/admin/settings', labelKey: 'nav.security', icon: ShieldCheck, perm: 'settings.manage' },
          { href: '/admin/orgs', labelKey: 'nav.orgs', icon: Building2, perm: 'org.manage' },
          { href: '/admin/users', labelKey: 'nav.users', icon: Users, perm: 'users.manage' },
          { href: '/admin/roles', labelKey: 'nav.roles', icon: ShieldHalf, perm: 'roles.manage' },
        ],
      },
    ],
  },
  {
    key: 'manager',
    labelKey: 'shell.sysManager',
    icon: Briefcase,
    adminOnly: false,
    groups: [
      {
        labelKey: 'shell.groupManager',
        items: [
          { href: '/manager/dashboard', labelKey: 'nav.managerDashboard', icon: LayoutDashboard, perm: 'manager.view' },
          { href: '/manager/profile', labelKey: 'nav.managerProfile', icon: User, perm: 'manager.view' },
        ],
      },
    ],
  },
  {
    key: 'user',
    labelKey: 'shell.sysUser',
    icon: UserCircle,
    adminOnly: false,
    groups: [
      {
        labelKey: 'shell.groupUser',
        items: [
          { href: '/user/dashboard', labelKey: 'nav.personalDashboard', icon: LayoutDashboard, perm: 'personal.view' },
          { href: '/user/profile', labelKey: 'nav.personalProfile', icon: User, perm: 'personal.view' },
        ],
      },
    ],
  },
];

// Систем тус бүрийн үндсэн зам — UserMenu-ийн профайл/тохиргооны холбоосыг
// идэвхтэй системд тааруулна.
const SYSTEM_BASE: Record<string, string> = { admin: '/admin', manager: '/manager', user: '/user' };

/**
 * Хоёр түвшний бүрхүүл — icon rail дахь "систем" (Admin / Personal) тус бүр
 * өөрийн бүлэгтэй дэд цэстэй. Admin системийг зөвхөн admin (role_id=1) харна;
 * энгийн хэрэглэгчид зөвхөн Personal систем харагдана. Бүх өнгө дизайн
 * системээс (OKLCH).
 */
export default function AppShell({ user, children }: Props) {
  const pathname = usePathname() ?? '/';
  const { T, lang } = useT();
  const isAdmin = user.roleId === ROLE_ADMIN;

  // Хэрэглэгчийн эрхүүдийг авч цэсийг шүүнэ. admin (role 1) бүгдийг харна;
  // ачаалж дуустал admin-д бүгд, бусдад зөвхөн эрхгүй (Personal) зүйл харагдана.
  const [perms, setPerms] = useState<string[] | null>(null);
  useEffect(() => {
    let alive = true;
    fetch('/api/rbac/me', { method: 'GET' })
      .then((r) => r.json())
      .then((b) => {
        if (alive && b?.ok && Array.isArray(b.data)) setPerms(b.data as string[]);
        else if (alive) setPerms([]);
      })
      .catch(() => alive && setPerms([]));
    return () => {
      alive = false;
    };
  }, []);

  const canSee = (perm?: string) => !perm || isAdmin || (perms?.includes(perm) ?? false);
  // Эрхтэй item-уудтай бүлэг/системийг л үлдээж шүүнэ.
  const visibleGroups = (s: NavSystem) =>
    s.groups
      .map((g) => ({ ...g, items: g.items.filter((i) => canSee(i.perm)) }))
      .filter((g) => g.items.length > 0);
  const systems = SYSTEMS.filter((s) => visibleGroups(s).length > 0);

  const isActive = (href: string) => (href === '/' ? pathname === '/' : pathname.startsWith(href));
  const systemMatches = (s: NavSystem) => visibleGroups(s).some((g) => g.items.some((i) => isActive(i.href)));

  // perms ачаалж дуустал admin биш хэрэглэгчид systems хоосон байж болзошгүй
  // тул activeSystem undefined байж болно — hook-уудыг найдвартай (optional)
  // дуудаж, дараа нь skeleton render хийнэ (undefined.key crash-аас сэргийлнэ).
  const activeSystem = systems.find(systemMatches) ?? systems[0];
  const [openKey, setOpenKey] = useState(activeSystem?.key ?? '');
  const [collapsed, setCollapsed] = useState(false);

  useEffect(() => {
    if (activeSystem) setOpenKey(activeSystem.key);
    // activeSystem.key-ээр л дахин ажиллана (object identity-ээр биш).
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeSystem?.key]);

  useEffect(() => {
    if (typeof window !== 'undefined' && window.innerWidth <= 900) setCollapsed(true);
  }, []);

  // Эрх ачаалж дуустал (эсвэл харагдах систем огт байхгүй үед) хөнгөн skeleton.
  if (!activeSystem) {
    return (
      <div className="shell2 shell2--loading" aria-busy="true">
        <aside className="iconrail" />
        <main className="maincol" />
      </div>
    );
  }

  const panel = systems.find((s) => s.key === openKey) ?? activeSystem;

  return (
    <div className={`shell2${collapsed ? ' is-collapsed' : ''}`}>
      {/* Нарийн icon rail — системүүд */}
      <aside className="iconrail">
        <Link href="/" className="iconrail__brand" aria-label="Gerege">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src="/brand.webp" alt="Gerege" />
        </Link>
        <nav className="iconrail__nav" aria-label={T('shell.navLabel')}>
          {systems.map((s) => {
            const Icon = s.icon;
            const active = s.key === activeSystem.key;
            return (
              <button
                key={s.key}
                type="button"
                className={`iconrail__btn${active ? ' is-active' : ''}`}
                title={T(s.labelKey)}
                aria-label={T(s.labelKey)}
                onClick={() => {
                  setOpenKey(s.key);
                  setCollapsed(false);
                }}
              >
                <Icon size={20} strokeWidth={2} />
              </button>
            );
          })}
        </nav>
        <div className="iconrail__bottom">
          <a
            className="iconrail__btn"
            href="https://gerege.mn/help"
            target="_blank"
            rel="noreferrer"
            title={T('nav.help')}
            aria-label={T('nav.help')}
          >
            <HelpCircle size={20} strokeWidth={2} />
          </a>
          <button
            className="iconrail__btn iconrail__signout"
            type="button"
            title={T('nav.signout')}
            aria-label={T('nav.signout')}
            onClick={() => signOut()}
          >
            <LogOut size={20} strokeWidth={2} />
          </button>
        </div>
      </aside>

      {/* Дэлгэгддэг дэд цэс — идэвхтэй системийн бүлгүүд (separator-тай) */}
      <aside className="sidepanel">
        <div className="sidepanel__head">
          <span className="sidepanel__brand-name">Gerege</span>
          <span className="sidepanel__title">{T(panel.labelKey)}</span>
        </div>
        <nav className="sidepanel__nav">
          {visibleGroups(panel).map((g, gi) => (
            <div key={gi} className="sidepanel__group">
              {g.labelKey && <span className="sidepanel__group-label">{T(g.labelKey)}</span>}
              {g.items.map((item) => {
                const Icon = item.icon;
                const active = isActive(item.href);
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={`sidepanel__link${active ? ' is-active' : ''}`}
                    aria-current={active ? 'page' : undefined}
                  >
                    <Icon size={16} strokeWidth={2} />
                    <span>{T(item.labelKey)}</span>
                  </Link>
                );
              })}
            </div>
          ))}
        </nav>
      </aside>

      {/* Контентын багана: header + scroll content */}
      <div className="maincol">
        <header className="topbar2">
          <button
            className="topbar2__toggle"
            type="button"
            aria-label={T('shell.toggleMenu')}
            onClick={() => setCollapsed((c) => !c)}
          >
            <Menu size={20} strokeWidth={2} />
          </button>
          <div className="topbar2__spacer" />
          <div className="topbar2__search">
            <Search size={16} strokeWidth={2} />
            <input className="topbar2__search-input" type="search" placeholder={T('shell.search')} aria-label={T('shell.search')} />
          </div>
          <div className="topbar2__actions">
            <UserMenu
              username={user.username}
              email={user.email}
              initials={user.initials}
              profileHref={`${SYSTEM_BASE[activeSystem.key] ?? '/admin'}/profile`}
              settingsHref={`${SYSTEM_BASE[activeSystem.key] ?? '/admin'}/settings`}
            />
          </div>
        </header>

        {lang === 'en' && <div className="i18n-banner">{T('banner.partial')}</div>}

        <main className="main">
          <div className="main__inner">{children}</div>
        </main>
      </div>
    </div>
  );
}
