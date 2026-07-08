"use client";

import Sidebar from "@/components/layout/Sidebar";
import TopHeader from "@/components/layout/TopHeader";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <>
      <Sidebar />
      <TopHeader />
      <main className="ml-[280px] pt-16 min-h-screen">
        <div className="max-w-[1440px] mx-auto p-gutter">{children}</div>
      </main>
    </>
  );
}
