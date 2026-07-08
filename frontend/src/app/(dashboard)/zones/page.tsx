"use client";

import { useEffect, useState } from "react";
import { getZones, ZoneInfo } from "@/lib/api";
import { clsx } from "clsx";

export default function ZonesPage() {
  const [zones, setZones] = useState<ZoneInfo[]>([]);
  const [summary, setSummary] = useState<{
    total_active: number;
    avg_multiplier: number;
    high_demand_peak: number;
  } | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<"all" | "active" | "peak">("all");
  const [toastMsg, setToastMsg] = useState<string | null>(null);

  useEffect(() => {
    async function load() {
      setLoading(true);
      setError(null);
      const res = await getZones();
      if (res.data) {
        setZones(res.data.zones);
        setSummary({
          total_active: res.data.total_active,
          avg_multiplier: res.data.avg_multiplier,
          high_demand_peak: res.data.high_demand_peak,
        });
      } else {
        setError(res.error || "Failed to load zones");
        // Fallback demo data
        setZones(getDemoZones());
        setSummary({
          total_active: 142,
          avg_multiplier: 1.45,
          high_demand_peak: 2.8,
        });
      }
      setLoading(false);
    }
    load();
  }, []);

  const filteredZones =
    filter === "all"
      ? zones
      : filter === "active"
        ? zones.filter((z) => z.status === "active")
        : zones.filter((z) => z.status === "peak" || z.demand === "high");

  function getUtilColor(pct: number): string {
    if (pct < 50) return "util-bar-green";
    if (pct <= 80) return "util-bar-amber";
    return "util-bar-red";
  }

  function getStatusBadge(status: string) {
    switch (status) {
      case "active":
        return "badge-success";
      case "peak":
        return "badge-error";
      case "inactive":
        return "badge-neutral";
      default:
        return "badge-neutral";
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-24">
        <svg
          className="animate-spin h-8 w-8 text-primary"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          />
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
          />
        </svg>
      </div>
    );
  }

  return (
    <div className="space-y-gutter">
      {/* KPI Cards */}
      {summary && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-gutter mb-xl">
          <div className="card-elevation rounded-xl p-lg border-l-4 border-primary group transition-all duration-300">
            <div className="flex justify-between items-start">
              <div>
                <p className="label-mono text-on-surface-variant uppercase mb-xs">
                  Active Zones
                </p>
                <h3
                  className="headline-lg text-on-surface"
                  style={{ fontFamily: "'JetBrains Mono', monospace" }}
                >
                  {summary.total_active}
                </h3>
              </div>
              <div className="bg-primary-fixed p-sm rounded-lg text-primary">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="24"
                  height="24"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" />
                  <circle cx="12" cy="10" r="3" />
                </svg>
              </div>
            </div>
          </div>

          <div className="card-elevation rounded-xl p-lg border-l-4 border-secondary-container group transition-all duration-300">
            <div className="flex justify-between items-start">
              <div>
                <p className="label-mono text-on-surface-variant uppercase mb-xs">
                  Avg. Multiplier
                </p>
                <h3
                  className="headline-lg text-on-surface"
                  style={{ fontFamily: "'JetBrains Mono', monospace" }}
                >
                  {summary.avg_multiplier.toFixed(2)}x
                </h3>
              </div>
              <div className="bg-secondary-fixed p-sm rounded-lg text-secondary">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="24"
                  height="24"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <polyline points="23 6 13.5 15.5 8.5 10.5 1 18" />
                  <polyline points="17 6 23 6 23 12" />
                </svg>
              </div>
            </div>
          </div>

          <div className="card-elevation rounded-xl p-lg border-l-4 border-tertiary group transition-all duration-300">
            <div className="flex justify-between items-start">
              <div>
                <p className="label-mono text-on-surface-variant uppercase mb-xs">
                  High Demand Peak
                </p>
                <h3
                  className="headline-lg text-on-surface"
                  style={{ fontFamily: "'JetBrains Mono', monospace" }}
                >
                  {summary.high_demand_peak.toFixed(1)}x
                </h3>
              </div>
              <div className="bg-tertiary-fixed p-sm rounded-lg text-tertiary">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="24"
                  height="24"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
                </svg>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Error Banner */}
      {error && (
        <div className="card-elevation rounded-xl p-lg border-l-4 border-error bg-error-container/30 mb-gutter">
          <p className="body-sm text-on-surface-variant">
            {error} — showing demo data.
          </p>
        </div>
      )}

      {/* Filters */}
      <div className="flex items-center justify-between mb-lg">
        <h4 className="title-md">Zone Directory</h4>
        <div className="flex bg-surface-container p-1 rounded-lg">
          {(["all", "active", "peak"] as const).map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={clsx(
                "px-md py-1 text-body-sm font-semibold rounded-md transition-colors",
                filter === f
                  ? "bg-surface-container-lowest shadow-sm text-primary"
                  : "text-on-surface-variant hover:text-on-surface"
              )}
            >
              {f === "all" ? "All Zones" : f === "active" ? "Active Only" : "High Demand"}
            </button>
          ))}
        </div>
      </div>

      {/* Zone Cards Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-gutter">
        {filteredZones.map((zone) => (
          <div
            key={zone.id}
            className="card-elevation rounded-xl overflow-hidden flex flex-col transition-all duration-300 hover:-translate-y-1 hover:shadow-md"
          >
            {/* Zone Image Placeholder */}
            <div className="relative h-36 bg-gradient-to-br from-surface-container-high to-surface-container flex items-center justify-center">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="48"
                height="48"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="1"
                strokeLinecap="round"
                strokeLinejoin="round"
                className="text-outline-variant/40"
              >
                <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" />
                <circle cx="12" cy="10" r="3" />
              </svg>
              <div className="absolute top-md right-md">
                <span className={clsx(getStatusBadge(zone.status), "uppercase")}>
                  {zone.status}
                </span>
              </div>
            </div>

            {/* Zone Info */}
            <div className="p-md flex-1">
              <div className="flex justify-between items-start mb-sm">
                <div>
                  <p className="label-mono text-on-surface-variant tracking-wider uppercase mb-0.5 text-[10px]">
                    {zone.code}
                  </p>
                  <h5 className="title-md text-on-surface">{zone.name}</h5>
                </div>
                {/* Toggle */}
                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    defaultChecked={zone.status !== "inactive"}
                    className="sr-only peer"
                    onChange={() => {
                      setToastMsg(`Zone ${zone.name} toggled`);
                      setTimeout(() => setToastMsg(null), 3000);
                    }}
                  />
                  <div className="w-10 h-5 bg-outline-variant rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-primary" />
                </label>
              </div>

              {/* Utilization Bar */}
              <div className="mb-sm">
                <p className="label-mono text-on-surface-variant uppercase mb-1 text-[10px]">
                  Utilization
                </p>
                <div className="h-2 w-full bg-surface-container-high rounded-full overflow-hidden">
                  <div
                    className={clsx(
                      "h-full rounded-full transition-all duration-500",
                      getUtilColor(zone.utilization)
                    )}
                    style={{ width: `${Math.min(zone.utilization, 100)}%` }}
                  />
                </div>
                <p
                  className={clsx(
                    "body-sm mt-1",
                    zone.utilization < 50
                      ? "text-secondary"
                      : zone.utilization <= 80
                        ? "text-tertiary"
                        : "text-error"
                  )}
                >
                  {zone.utilization}% —{" "}
                  {zone.utilization < 50
                    ? "Low"
                    : zone.utilization <= 80
                      ? "Moderate"
                      : "High Utilization"}
                </p>
              </div>

              {/* Multiplier & Demand */}
              <div className="grid grid-cols-2 gap-sm pt-sm border-t border-outline-variant/30">
                <div>
                  <p className="label-mono text-on-surface-variant uppercase text-[10px]">
                    Multiplier
                  </p>
                  <p className="title-md text-primary font-bold">
                    {zone.multiplier.toFixed(1)}x
                  </p>
                </div>
                <div className="text-right">
                  <p className="label-mono text-on-surface-variant uppercase text-[10px]">
                    Demand
                  </p>
                  <p
                    className={clsx(
                      "title-md font-bold",
                      zone.demand === "high"
                        ? "text-error"
                        : zone.demand === "medium"
                          ? "text-secondary"
                          : "text-on-surface-variant"
                    )}
                  >
                    {zone.demand.charAt(0).toUpperCase() + zone.demand.slice(1)}
                  </p>
                </div>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="p-sm bg-surface-container-low flex justify-end gap-xs border-t border-outline-variant/10">
              <button className="p-xs text-on-surface-variant hover:text-primary hover:bg-surface-container-high rounded-lg transition-all">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="18"
                  height="18"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
                  <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
                </svg>
              </button>
              <button className="p-xs text-on-surface-variant hover:text-secondary hover:bg-surface-container-high rounded-lg transition-all">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="18"
                  height="18"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <path d="M18 20V10" />
                  <path d="M12 20V4" />
                  <path d="M6 20v-6" />
                </svg>
              </button>
            </div>
          </div>
        ))}

        {/* Add Zone Placeholder Card */}
        <div className="border-2 border-dashed border-outline-variant rounded-xl flex flex-col items-center justify-center p-xl group hover:border-primary transition-all cursor-pointer min-h-[320px]">
          <div className="w-12 h-12 rounded-full bg-surface-container flex items-center justify-center text-on-surface-variant group-hover:bg-primary-fixed group-hover:text-primary transition-all mb-md">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="32"
              height="32"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <line x1="12" y1="5" x2="12" y2="19" />
              <line x1="5" y1="12" x2="19" y2="12" />
            </svg>
          </div>
          <p className="title-md text-on-surface-variant group-hover:text-primary transition-all">
            Create New Zone
          </p>
          <p className="body-sm text-on-surface-variant opacity-60 text-center mt-xs">
            Define new boundaries and pricing rules
          </p>
        </div>
      </div>

      {/* Toast */}
      {toastMsg && (
        <div className="fixed bottom-xl left-1/2 -translate-x-1/2 bg-secondary-container text-on-secondary-container px-lg py-sm rounded-xl shadow-lg z-50 toast-enter flex items-center gap-md">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
            <polyline points="22 4 12 14.01 9 11.01" />
          </svg>
          <span className="button-text">{toastMsg}</span>
        </div>
      )}
    </div>
  );
}

/* Fallback demo data when API is unavailable */
function getDemoZones(): ZoneInfo[] {
  return [
    {
      id: "zone-sj-01",
      name: "South Jakarta",
      code: "ZONE-SJ-01",
      utilization: 78,
      status: "active",
      multiplier: 1.2,
      demand: "medium",
      surge_threshold: 85,
    },
    {
      id: "zone-ch-04",
      name: "Central Hub",
      code: "ZONE-CH-04",
      utilization: 94,
      status: "peak",
      multiplier: 2.4,
      demand: "high",
      surge_threshold: 80,
    },
    {
      id: "zone-wh-12",
      name: "West Harbor",
      code: "ZONE-WH-12",
      utilization: 22,
      status: "inactive",
      multiplier: 1.0,
      demand: "low",
      surge_threshold: 90,
    },
    {
      id: "zone-ni-09",
      name: "North Industrial",
      code: "ZONE-NI-09",
      utilization: 45,
      status: "active",
      multiplier: 1.1,
      demand: "low",
      surge_threshold: 85,
    },
    {
      id: "zone-ap-03",
      name: "Airport Terminal 1",
      code: "ZONE-AP-03",
      utilization: 88,
      status: "peak",
      multiplier: 1.8,
      demand: "high",
      surge_threshold: 75,
    },
    {
      id: "zone-cbd-07",
      name: "Downtown Core",
      code: "ZONE-CBD-07",
      utilization: 65,
      status: "active",
      multiplier: 1.5,
      demand: "medium",
      surge_threshold: 80,
    },
    {
      id: "zone-er-15",
      name: "East Riverside",
      code: "ZONE-ER-15",
      utilization: 35,
      status: "active",
      multiplier: 1.05,
      demand: "low",
      surge_threshold: 90,
    },
    {
      id: "zone-sp-22",
      name: "South Port",
      code: "ZONE-SP-22",
      utilization: 52,
      status: "active",
      multiplier: 1.3,
      demand: "medium",
      surge_threshold: 85,
    },
  ];
}
