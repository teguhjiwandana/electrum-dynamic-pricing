"use client";

import { useState, useEffect, FormEvent } from "react";
import { calculatePricing, getZones, PricingResult, PricingParams, ZoneInfo } from "@/lib/api";
import { clsx } from "clsx";

export default function PricingPage() {
  const [vehicleId, setVehicleId] = useState("EV-10001");
  const [zone, setZone] = useState("south-jakarta");
  const [duration, setDuration] = useState("3");
  const [result, setResult] = useState<PricingResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [zones, setZones] = useState<ZoneInfo[]>([]);

  useEffect(() => {
    getZones().then((res) => {
      if (res.data) setZones(res.data);
    });
  }, []);

  async function handleCalculate(e?: FormEvent) {
    if (e) e.preventDefault();
    setError(null);
    setLoading(true);

    const params: PricingParams = {
      vehicle_id: vehicleId,
      zone,
      duration_hours: parseInt(duration) || 1,
    };

    const res = await calculatePricing(params);

    if (res.data) {
      setResult(res.data);
      setError(null);
    } else {
      setResult(null);
      setError(res.error || "Failed to calculate price");
    }
    setLoading(false);
  }

  return (
    <div className="space-y-gutter">
      {/* Header */}
      <section className="mb-xl">
        <div className="flex justify-between items-end">
          <div>
            <p className="label-mono text-secondary uppercase mb-xs">
              Calculator Workspace
            </p>
            <h2 className="headline-lg text-primary">Pricing Simulation</h2>
          </div>
        </div>
      </section>

      {/* Metric Tiles */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-gutter mb-gutter">
        <div className="card-elevation rounded-xl p-lg metric-accent-primary">
          <p className="label-mono text-on-surface-variant uppercase mb-xs">
            Base Rate
          </p>
          <p className="headline-lg text-primary">Rp 6.250</p>
          <p className="body-sm text-on-surface-variant mt-xs">Per hour</p>
        </div>
        <div className="card-elevation rounded-xl p-lg metric-accent-secondary">
          <p className="label-mono text-on-surface-variant uppercase mb-xs">
            Surge Cap
          </p>
          <p className="headline-lg text-primary">2.0×</p>
          <p className="body-sm text-on-surface-variant mt-xs">Max multiplier</p>
        </div>
        <div className="card-elevation rounded-xl p-lg metric-accent-tertiary">
          <p className="label-mono text-on-surface-variant uppercase mb-xs">
            Active Zones
          </p>
          <p className="headline-lg text-primary">{zones.length || "—"}</p>
          <p className="body-sm text-on-surface-variant mt-xs">
            Across Jakarta
          </p>
        </div>
      </div>

      {/* Main Split */}
      <div className="grid grid-cols-12 gap-gutter items-start">
        {/* Left: Input Form */}
        <div className="col-span-12 lg:col-span-5 space-y-gutter">
          <div className="card-elevation rounded-xl p-xl">
            <div className="flex items-center gap-sm mb-lg">
              <span className="p-sm rounded-lg bg-primary-fixed text-primary">
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <polyline points="9 10 4 15 9 20" />
                  <path d="M20 4v7a4 4 0 0 1-4 4H4" />
                </svg>
              </span>
              <h3 className="title-md">Pricing Parameters</h3>
            </div>

            <form onSubmit={handleCalculate} className="space-y-md">
              {/* Vehicle ID */}
              <div>
                <label htmlFor="vehicleId" className="label-mono text-on-surface-variant uppercase block mb-xs">
                  Vehicle ID
                </label>
                <input id="vehicleId" type="text" value={vehicleId} onChange={(e) => setVehicleId(e.target.value)} className="input-field" placeholder="EV-10001" />
              </div>

              {/* Zone */}
              <div>
                <label htmlFor="zone" className="label-mono text-on-surface-variant uppercase block mb-xs">
                  Zone Selection
                </label>
                <select id="zone" value={zone} onChange={(e) => setZone(e.target.value)} className="input-field bg-white">
                  {zones.length === 0 && (
                    <option value="south-jakarta">Loading zones...</option>
                  )}
                  {zones.map((z) => (
                    <option key={z.code} value={z.code}>
                      {z.name} ({Math.round(z.utilization)}%)
                    </option>
                  ))}
                </select>
              </div>

              {/* Duration */}
              <div>
                <label htmlFor="duration" className="label-mono text-on-surface-variant uppercase block mb-xs">
                  Duration (Hours)
                </label>
                <input id="duration" type="number" min="1" step="1" value={duration} onChange={(e) => setDuration(e.target.value)} className="input-field" placeholder="3" />
              </div>

              {/* Calculate Button */}
              <button type="submit" disabled={loading} className={clsx("btn-primary w-full py-md text-lg font-semibold mt-xl")}>
                {loading ? "Calculating..." : "Calculate Price"}
              </button>
            </form>
          </div>

          {/* Info Card */}
          <div className="card-elevation rounded-xl p-lg bg-surface-container-high border-none flex gap-md">
            <svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-tertiary shrink-0 mt-0.5">
              <circle cx="12" cy="12" r="10" /><line x1="12" y1="16" x2="12" y2="12" /><line x1="12" y1="8" x2="12.01" y2="8" />
            </svg>
            <div>
              <p className="button-text text-on-surface">Real-time Factors</p>
              <p className="body-sm text-on-surface-variant">Demand, zone surge, and battery discount are based on current configuration and fleet data.</p>
            </div>
          </div>
        </div>

        {/* Right: Results */}
        <div className="col-span-12 lg:col-span-7 space-y-gutter">
          {/* Error State */}
          {error && (
            <div className="card-elevation rounded-xl p-lg border-l-4 border-error bg-error-container/30">
              <p className="body-sm text-on-surface-variant">{error}</p>
            </div>
          )}

          {/* Results Card */}
          {result && (
            <>
              <div className="card-elevation rounded-xl p-xl">
                <div className="flex justify-between items-center mb-lg">
                  <h3 className="title-md">Price Breakdown</h3>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-lg">
                  {/* Base Rate */}
                  <div className="card-elevation rounded-lg p-md metric-accent-primary flex justify-between items-center">
                    <div>
                      <p className="label-mono text-on-surface-variant mb-base uppercase">Base Rate</p>
                      <p className="headline-lg">Rp {result.breakdown.base_rate_per_hour.toLocaleString()}<span className="body-sm text-on-surface-variant font-normal">/hr</span></p>
                    </div>
                  </div>

                  {/* Demand Multiplier */}
                  <div className="card-elevation rounded-lg p-md metric-accent-secondary flex justify-between items-center">
                    <div>
                      <p className="label-mono text-on-surface-variant mb-base uppercase">Demand Multiplier</p>
                      <p className="headline-lg">{result.breakdown.demand_multiplier.toFixed(2)}×</p>
                    </div>
                  </div>

                  {/* Zone Surge */}
                  <div className="card-elevation rounded-lg p-md metric-accent-tertiary flex justify-between items-center">
                    <div>
                      <p className="label-mono text-on-surface-variant mb-base uppercase">Zone Surge</p>
                      <p className="headline-lg">{result.breakdown.zone_surge_factor.toFixed(2)}×</p>
                    </div>
                  </div>

                  {/* Battery Discount */}
                  <div className="card-elevation rounded-lg p-md metric-accent-secondary flex justify-between items-center">
                    <div>
                      <p className="label-mono text-on-surface-variant mb-base uppercase">Battery Discount</p>
                      <p className="headline-lg text-secondary">{(result.breakdown.battery_discount_factor * 100).toFixed(0)}%</p>
                    </div>
                  </div>
                </div>

                {/* Grand Total */}
                <div className="mt-xl pt-xl border-t border-outline-variant flex items-center justify-between">
                  <div>
                    <p className="body-md text-on-surface-variant">Total Estimated Cost</p>
                    <p className="display-lg text-primary" style={{ fontFamily: "'JetBrains Mono', monospace" }}>
                      Rp {result.total_price.toLocaleString()}
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="label-mono text-on-surface-variant mb-xs uppercase">Duration</p>
                    <p className="title-md text-secondary">{result.duration_hours} hours</p>
                  </div>
                </div>
              </div>
            </>
          )}

          {/* Empty State */}
          {!result && !error && !loading && (
            <div className="card-elevation rounded-xl p-xl flex flex-col items-center justify-center py-16 text-center">
              <h3 className="title-md text-on-surface-variant mb-xs">No Calculation Yet</h3>
              <p className="body-sm text-on-surface-variant max-w-sm">
                Enter pricing parameters and click Calculate to see results with detailed cost breakdown.
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
