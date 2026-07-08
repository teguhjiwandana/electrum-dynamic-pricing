"use client";

import { usePathname } from "next/navigation";
import { clsx } from "clsx";
import {
  LayoutDashboard,
  Calculator,
  History,
  Settings,
  Globe,
} from "lucide-react";
import Link from "next/link";

const navItems = [
  { href: "/pricing", label: "Pricing Calculator", icon: Calculator },
  { href: "/audit", label: "Audit Log", icon: History },
  { href: "/zones", label: "Zones", icon: Globe },
  { href: "/settings", label: "Settings", icon: Settings },
];

export default function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="w-[280px] h-screen fixed left-0 top-0 bg-surface-container-lowest flex flex-col py-lg px-md z-50 border-r border-outline-variant/20 shadow-sm">
      {/* Brand Header */}
      <div className="mb-xl px-xs">
        <Link
          href="/pricing"
          className="headline-lg font-extrabold text-primary no-underline tracking-tight"
        >
          Electrum DPE
        </Link>
        <p className="body-sm text-on-surface-variant opacity-70 mt-xs">
          Enterprise Admin
        </p>
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-xs custom-scrollbar overflow-y-auto">
        <Link
          href="/pricing"
          className={clsx(
            "flex items-center gap-md px-md py-sm rounded-lg transition-all duration-200",
            pathname === "/pricing"
              ? "text-on-primary bg-primary-container font-semibold"
              : "text-on-surface-variant hover:bg-surface-container-high"
          )}
        >
          <LayoutDashboard size={20} />
          <span className="body-md">Dashboard</span>
        </Link>

        {navItems.map((item) => {
          const isActive =
            pathname === item.href || pathname.startsWith(item.href + "/");
          const Icon = item.icon;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={clsx(
                "flex items-center gap-md px-md py-sm rounded-lg transition-all duration-200",
                isActive
                  ? "text-on-primary bg-primary-container font-semibold shadow-sm"
                  : "text-on-surface-variant hover:bg-surface-container-high"
              )}
            >
              <Icon size={20} />
              <span className="body-md">{item.label}</span>
            </Link>
          );
        })}
      </nav>

      {/* User Profile */}
      <div className="mt-auto pt-lg border-t border-outline-variant/30 flex items-center gap-md px-xs">
        <div className="w-10 h-10 rounded-full bg-primary-fixed flex items-center justify-center shrink-0">
          <span className="label-mono text-primary font-bold">MC</span>
        </div>
        <div className="flex flex-col min-w-0">
          <span className="button-text text-on-surface truncate">
            Marcus Chen
          </span>
          <span className="body-sm text-on-surface-variant truncate">
            Senior Analyst
          </span>
        </div>
      </div>
    </aside>
  );
}
