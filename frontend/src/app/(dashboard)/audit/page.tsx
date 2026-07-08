"use client";

import { useEffect, useState, useCallback } from "react";
import { getAuditLogs, getZones, getVehicles, AuditLogEntry, AuditFactors, ZoneInfo, VehicleInfo } from "@/lib/api";
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

  const [zones, setZones] = useState<ZoneInfo[]>([]);
  const [vehicles, setVehicles] = useState<VehicleInfo[]>([]);

  useEffect(() => {
    getZones().then((res) => { if (res.data) setZones(res.data); }).catch(() => {});
    getVehicles().then((res) => { if (res.data) setVehicles(res.data); }).catch(() => {});
  }, []);

  const fetchLogs = useCallback(async () => {
    setLoading(true);
    setError(null);

    const filters: { vehicle_id?: string; zone?: string } = {};
    if (filterVehicle.trim()) filters.vehicle_id = filterVehicle.trim();
    if (filterZone.trim()) filters.zone = filterZone.trim();

    const res = await getAuditLogs(page, pageSize, filters);

    if (res.data) {
      setEntries(res.data.data);
      setTotal(res.data.total);
      setTotalPages(res.data.total_pages);
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

  return (
    <div className="space-y-gutter">
      {/* Header */}
      <div className="flex items-end justify-between mb-xl">
        <div>
          <h2 className="headline-lg text-on-surface">Audit Log</h2>
          <p className="body-md text-on-surface-variant">
            All pricing calculations with tamper-evident HMAC signatures.
          </p>
        </div>
      </div>

      {/* Filters */}
      <div className="card-elevation rounded-xl p-lg flex flex-wrap items-center gap-md">
        <label className="label-mono text-on-surface-variant uppercase">Filter:</label>
        <select
          value={filterVehicle}
          onChange={(e) => { setFilterVehicle(e.target.value); setPage(1); }}
          className="input-field bg-white min-w-[200px]"
        >
          <option value="">All Vehicles</option>
          {vehicles.map((v) => (
            <option key={v.id} value={v.id}>{v.id} — {v.model}</option>
          ))}
        </select>
        <select
          value={filterZone}
          onChange={(e) => { setFilterZone(e.target.value); setPage(1); }}
          className="input-field bg-white min-w-[200px]"
        >
          <option value="">All Zones</option>
          {zones.map((z) => (
            <option key={z.code} value={z.code}>{z.name}</option>
          ))}
        </select>
        <button onClick={fetchLogs} className="btn-secondary px-lg py-sm">
          Refresh
        </button>
      </div>

      {/* Error */}
      {error && (
        <div className="card-elevation rounded-xl p-lg border-l-4 border-error bg-error-container/30">
          <p className="body-sm text-on-surface-variant">{error}</p>
        </div>
      )}

      {/* Loading */}
      {loading && (
        <div className="flex items-center justify-center py-24">
          <svg className="animate-spin h-8 w-8 text-primary" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
        </div>
      )}

      {/* Empty */}
      {!loading && !error && entries.length === 0 && (
        <div className="card-elevation rounded-xl p-xl flex flex-col items-center justify-center py-16 text-center">
          <h3 className="title-md text-on-surface-variant mb-xs">No Audit Entries</h3>
          <p className="body-sm text-on-surface-variant max-w-sm">
            Pricing calculations will appear here with their full breakdown and HMAC signature.
          </p>
        </div>
      )}

      {/* Table */}
      {!loading && entries.length > 0 && (
        <div className="card-elevation rounded-xl overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="bg-surface-container-low">
                  <th className="label-mono uppercase text-on-surface-variant p-md text-left">Timestamp</th>
                  <th className="label-mono uppercase text-on-surface-variant p-md text-left">Vehicle ID</th>
                  <th className="label-mono uppercase text-on-surface-variant p-md text-left">Zone</th>
                  <th className="label-mono uppercase text-on-surface-variant p-md text-left">Duration</th>
                  <th className="label-mono uppercase text-on-surface-variant p-md text-right">Price (Rp)</th>
                  <th className="label-mono uppercase text-on-surface-variant p-md text-left">Factors</th>
                  <th className="label-mono uppercase text-on-surface-variant p-md text-left">Signature</th>
                </tr>
              </thead>
              <tbody>
                {entries.map((entry, idx) => (
                  <tr key={entry.id || idx} className={clsx("border-t border-outline-variant/30", idx % 2 === 1 && "bg-surface-container-low/30")}>
                    <td className="p-md label-mono text-on-surface text-xs">{formatTimestamp(entry.timestamp)}</td>
                    <td className="p-md body-sm text-on-surface">{entry.vehicle_id}</td>
                    <td className="p-md body-sm text-on-surface">{entry.zone}</td>
                    <td className="p-md body-sm text-on-surface">{entry.duration_hours}h</td>
                    <td className="p-md label-mono text-right font-semibold">Rp {entry.final_price.toLocaleString()}</td>
                    <td className="p-md body-sm text-on-surface-variant text-xs">
                      base: {entry.factors_applied?.base_rate_per_hour}
                      , demand: {entry.factors_applied?.demand_multiplier}×
                      , surge: {entry.factors_applied?.zone_surge_factor}×
                      , batt: {entry.factors_applied?.battery_discount_factor}
                    </td>
                    <td className="p-md label-mono text-on-surface-variant text-xs" style={{ maxWidth: 120 }}>
                      {entry.signature?.slice(0, 16)}...
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          <div className="flex items-center justify-between p-lg border-t border-outline-variant/30 bg-surface-container-low/50">
            <span className="body-sm text-on-surface-variant">
              {total} entries • Page {page} of {totalPages}
            </span>
            <div className="flex gap-sm">
              <button
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                className={clsx("btn-secondary px-md py-sm text-sm", page <= 1 && "opacity-50")}
              >
                Previous
              </button>
              <button
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
                className={clsx("btn-secondary px-md py-sm text-sm", page >= totalPages && "opacity-50")}
              >
                Next
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
