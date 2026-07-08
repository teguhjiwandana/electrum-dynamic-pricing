"use client";

import { usePathname } from "next/navigation";

const pageTitles: Record<string, string> = {
  "/pricing": "Pricing Calculator",
  "/audit": "Audit Log",
  "/zones": "Zones Management",
  "/settings": "Settings",
};

export default function TopHeader() {
  const pathname = usePathname();
  const title = pageTitles[pathname] || "Dynamic Pricing Engine";

  return (
    <header className="fixed top-0 right-0 left-[280px] h-16 bg-surface-container-lowest shadow-sm z-40 px-lg flex items-center justify-between">
      <h1 className="title-md font-bold text-on-surface truncate">{title}</h1>

      <div className="flex items-center gap-md">
        <button className="p-2 rounded-full text-on-surface-variant hover:bg-surface-container-high transition-colors">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" />
            <path d="M13.73 21a2 2 0 0 1-3.46 0" />
          </svg>
        </button>
        <button className="p-2 rounded-full text-on-surface-variant hover:bg-surface-container-high transition-colors">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <circle cx="12" cy="12" r="10" />
            <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
            <line x1="12" y1="17" x2="12.01" y2="17" />
          </svg>
        </button>
      </div>
    </header>
  );
}
