"use client";

import { useEffect, useState, useCallback } from "react";
import { getAuditLogs, AuditLogEntry } from "@/lib/api";
import { clsx } from "clsx";

export default function AuditPage() {
  const [entries, setEntries] = useState<AuditLogEntry[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filterVehicle, setFilterVehicle] = useState("");
  const [filterZone, setFilterZone] = useState("");
  const pageSize = 20;

  const fetchLogs = useCallback(async () => {
    setLoading(true);
    setError(null);

    const filters: { vehicle_id?: string; zone?: string } = {};
    if (filterVehicle.trim()) filters.vehicle_id = filterVehicle.trim();
    if (filterZone.trim()) filters.zone = filterZone.trim();

    const res = await getAuditLogs(page, pageSize, filters);

    if (res.data) {
      setEntries(res.data.entries);
      setTotal(res.data.total);
      setTotalPages(Math.ceil(res.data.total / pageSize));
    } else {
      setError(res.error || "Failed to load audit logs");
      setEntries([]);
    }
    setLoading(false);
  }, [page, filterVehicle, filterZone]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  function formatTimestamp(ts: string): string {
    try {
      const d = new Date(ts);
      return d.toISOString().replace("T", " ").slice(0, 19);
    } catch {
      return ts;
    }
  }

  function formatFactors(factors: Record<string, number>): string {
    return Object.entries(factors)
      .map(([k, v]) => `${k}: ${v}`)
      .join(", ");
  }

  return (
    <div className="space-y-gutter">
      {/* Header */}
      <div className="flex items-end justify-between mb-xl">
        <div>
          <h2 className="headline-lg text-on-surface">System Audit Logs</h2>
          <p className="body-md text-on-surface-variant">
            Track all system activities and pricing adjustments in real-time.
          </p>
        </div>
        <div className="flex gap-md">
          <button className="btn-secondary px-lg py-sm flex items-center gap-xs">
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
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" y1="15" x2="12" y2="3" />
            </svg>
            Export Report
          </button>
        </div>
      </div>

      {/* Filters */}
      <div className="card-elevation rounded-xl p-lg flex flex-wrap items-center gap-md">
        <div className="flex items-center gap-sm">
          <label className="label-mono text-on-surface-variant uppercase">
            Filter:
          </label>
          <input
            type="text"
            placeholder="Vehicle ID"
            value={filterVehicle}
            onChange={(e) => {
              setFilterVehicle(e.target.value);
              setPage(1);
            }}
            className="input-field max-w-[180px] text-sm"
          />
          <input
            type="text"
            placeholder="Zone"
            value={filterZone}
            onChange={(e) => {
              setFilterZone(e.target.value);
              setPage(1);
            }}
            className="input-field max-w-[180px] text-sm"
          />
        </div>
        <button
          onClick={fetchLogs}
          className="btn-secondary px-md py-xs text-sm ml-auto"
          disabled={loading}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            className={clsx(loading && "animate-spin")}
          >
            <polyline points="23 4 23 10 17 10" />
            <polyline points="1 20 1 14 7 14" />
            <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" />
          </svg>
          Refresh
        </button>
      </div>

      {/* Error State */}
      {error && (
        <div className="card-elevation rounded-xl p-lg border-l-4 border-error bg-error-container/30">
          <p className="body-sm text-error">{error}</p>
        </div>
      )}

      {/* Loading State */}
      {loading && (
        <div className="card-elevation rounded-xl p-xl flex items-center justify-center gap-md py-16">
          <svg
            className="animate-spin h-6 w-6 text-primary"
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
          <span className="body-md text-on-surface-variant">
            Loading audit logs...
          </span>
        </div>
      )}

      {/* Empty State */}
      {!loading && !error && entries.length === 0 && (
        <div className="card-elevation rounded-xl p-xl flex flex-col items-center justify-center py-16 text-center">
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
            className="text-outline-variant mb-lg"
          >
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
            <polyline points="14 2 14 8 20 8" />
            <line x1="16" y1="13" x2="8" y2="13" />
            <line x1="16" y1="17" x2="8" y2="17" />
          </svg>
          <h3 className="title-md text-on-surface-variant mb-xs">
            No Audit Logs Found
          </h3>
          <p className="body-sm text-on-surface-variant max-w-sm">
            No log entries match your current filters. Try adjusting the search
            criteria.
          </p>
        </div>
      )}

      {/* Table */}
      {!loading && entries.length > 0 && (
        <div className="card-elevation rounded-xl overflow-hidden">
          <div className="overflow-x-auto custom-scrollbar">
            <table className="w-full text-left border-collapse">
              <thead className="bg-surface-container-lowest sticky top-0 z-10">
                <tr className="border-b border-outline-variant/20">
                  <th className="px-lg py-md label-mono uppercase text-on-surface-variant">
                    Timestamp
                  </th>
                  <th className="px-lg py-md label-mono uppercase text-on-surface-variant">
                    Vehicle ID
                  </th>
                  <th className="px-lg py-md label-mono uppercase text-on-surface-variant">
                    Zone
                  </th>
                  <th className="px-lg py-md label-mono uppercase text-on-surface-variant">
                    Duration
                  </th>
                  <th className="px-lg py-md label-mono uppercase text-on-surface-variant text-right">
                    Final Price
                  </th>
                  <th className="px-lg py-md label-mono uppercase text-on-surface-variant">
                    Factors
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-outline-variant/10">
                {entries.map((entry, idx) => (
                  <tr
                    key={entry.id || idx}
                    className={clsx(
                      "transition-colors group",
                      idx % 2 === 1 ? "bg-surface-container-low/30" : "bg-surface-container-lowest"
                    )}
                  >
                    <td
                      className="px-lg py-md label-mono text-on-surface-variant whitespace-nowrap"
                      style={{ fontSize: "12px" }}
                    >
                      {formatTimestamp(entry.timestamp)}
                    </td>
                    <td className="px-lg py-md">
                      <span
                        className="label-mono text-on-surface"
                        style={{ fontSize: "12px" }}
                      >
                        {entry.vehicle_id}
                      </span>
                    </td>
                    <td className="px-lg py-md">
                      <span className="body-sm text-on-surface">
                        {entry.zone}
                      </span>
                    </td>
                    <td className="px-lg py-md">
                      <span className="body-sm text-on-surface">
                        {entry.duration_hours}h
                      </span>
                    </td>
                    <td className="px-lg py-md text-right">
                      <span
                        className="label-mono text-primary font-bold"
                        style={{ fontSize: "14px" }}
                      >
                        ${entry.final_price.toFixed(2)}
                      </span>
                    </td>
                    <td className="px-lg py-md">
                      <span className="body-sm text-on-surface-variant max-w-[200px] truncate block">
                        {formatFactors(entry.factors)}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          <div className="px-lg py-md bg-surface-container-low border-t border-outline-variant/10 flex items-center justify-between">
            <span className="body-sm text-on-surface-variant">
              Showing {entries.length} of {total} entries
            </span>
            <div className="flex items-center gap-xs">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="p-xs rounded-lg hover:bg-surface-container-highest transition-colors disabled:opacity-30"
              >
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
                  <polyline points="15 18 9 12 15 6" />
                </svg>
              </button>

              {Array.from({ length: Math.min(totalPages, 5) }, (_, i) => {
                const startPage = Math.max(
                  1,
                  Math.min(page - 2, totalPages - 4)
                );
                const pageNum = startPage + i;
                if (pageNum > totalPages) return null;
                return (
                  <button
                    key={pageNum}
                    onClick={() => setPage(pageNum)}
                    className={clsx(
                      "px-md py-xs rounded font-button-text text-body-sm transition-colors",
                      pageNum === page
                        ? "bg-primary-container text-on-primary"
                        : "text-on-surface-variant hover:bg-surface-container-highest"
                    )}
                  >
                    {pageNum}
                  </button>
                );
              })}

              <button
                onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                disabled={page === totalPages}
                className="p-xs rounded-lg hover:bg-surface-container-highest transition-colors disabled:opacity-30"
              >
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
                  <polyline points="9 18 15 12 9 6" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
