"use client";

import { useState, FormEvent } from "react";
import { calculatePricing, PricingResult, PricingParams } from "@/lib/api";
import { clsx } from "clsx";

export default function PricingPage() {
  const [vehicleId, setVehicleId] = useState("EV-7729-ALPHA");
  const [zone, setZone] = useState("central-business-district");
  const [duration, setDuration] = useState("12");
  const [result, setResult] = useState<PricingResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleCalculate(e?: FormEvent) {
    if (e) e.preventDefault();
    setError(null);
    setLoading(true);

    const params: PricingParams = {
      vehicle_id: vehicleId,
      zone,
      duration_hours: parseFloat(duration) || 0,
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
            <h2 className="headline-lg text-primary">Simulation &amp; Estimation</h2>
          </div>
          <div className="flex gap-md">
            <button
              className="btn-secondary px-lg py-sm"
              onClick={() => handleCalculate()}
              disabled={loading}
            >
              New Simulation
            </button>
          </div>
        </div>
      </section>

      {/* Metric Tiles */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-gutter mb-gutter">
        <div className="card-elevation rounded-xl p-lg metric-accent-primary">
          <p className="label-mono text-on-surface-variant uppercase mb-xs">
            Config Version
          </p>
          <p className="headline-lg text-primary">v4.2.0</p>
          <p className="body-sm text-on-surface-variant mt-xs">
            Stable release
          </p>
        </div>
        <div className="card-elevation rounded-xl p-lg metric-accent-secondary">
          <p className="label-mono text-on-surface-variant uppercase mb-xs">
            Base Rate
          </p>
          <p className="headline-lg text-primary">$14.50</p>
          <p className="body-sm text-on-surface-variant mt-xs">Per hour</p>
        </div>
        <div className="card-elevation rounded-xl p-lg metric-accent-tertiary">
          <p className="label-mono text-on-surface-variant uppercase mb-xs">
            Active Zones
          </p>
          <p className="headline-lg text-primary">142</p>
          <p className="body-sm text-on-surface-variant mt-xs">
            Across all regions
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
                  <polyline points="9 10 4 15 9 20" />
                  <path d="M20 4v7a4 4 0 0 1-4 4H4" />
                </svg>
              </span>
              <h3 className="title-md">Pricing Parameters</h3>
            </div>

            <form onSubmit={handleCalculate} className="space-y-md">
              {/* Vehicle ID */}
              <div>
                <label
                  htmlFor="vehicleId"
                  className="label-mono text-on-surface-variant uppercase block mb-xs"
                >
                  Vehicle ID
                </label>
                <input
                  id="vehicleId"
                  type="text"
                  value={vehicleId}
                  onChange={(e) => setVehicleId(e.target.value)}
                  className="input-field"
                  placeholder="EV-7729-ALPHA"
                />
              </div>

              {/* Zone */}
              <div>
                <label
                  htmlFor="zone"
                  className="label-mono text-on-surface-variant uppercase block mb-xs"
                >
                  Zone Selection
                </label>
                <select
                  id="zone"
                  value={zone}
                  onChange={(e) => setZone(e.target.value)}
                  className="input-field bg-white"
                >
                  <option value="central-business-district">
                    Central Business District (High Traffic)
                  </option>
                  <option value="peripheral-residential">
                    Peripheral Residential
                  </option>
                  <option value="industrial-hub-b">Industrial Hub B</option>
                  <option value="airport-logistics">
                    Airport Logistics Zone
                  </option>
                </select>
              </div>

              {/* Duration */}
              <div>
                <label
                  htmlFor="duration"
                  className="label-mono text-on-surface-variant uppercase block mb-xs"
                >
                  Duration (Hours)
                </label>
                <input
                  id="duration"
                  type="number"
                  min="0.5"
                  step="0.5"
                  value={duration}
                  onChange={(e) => setDuration(e.target.value)}
                  className="input-field"
                  placeholder="12"
                />
              </div>

              {/* Calculate Button */}
              <button
                type="submit"
                disabled={loading}
                className={clsx(
                  "btn-primary w-full py-md text-lg font-semibold mt-xl"
                )}
              >
                {loading ? (
                  <>
                    <svg
                      className="animate-spin h-5 w-5"
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
                    Processing...
                  </>
                ) : (
                  <>
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
                      <path d="M18 20V10" />
                      <path d="M12 20V4" />
                      <path d="M6 20v-6" />
                    </svg>
                    Calculate Price
                  </>
                )}
              </button>
            </form>
          </div>

          {/* Info Card */}
          <div className="card-elevation rounded-xl p-lg bg-surface-container-high border-none flex gap-md">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="22"
              height="22"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
              className="text-tertiary shrink-0 mt-0.5"
            >
              <circle cx="12" cy="12" r="10" />
              <line x1="12" y1="16" x2="12" y2="12" />
              <line x1="12" y1="8" x2="12.01" y2="8" />
            </svg>
            <div>
              <p className="button-text text-on-surface">Live Data Sync</p>
              <p className="body-sm text-on-surface-variant">
                Factors are based on live market demand. Calculations may vary
                with real-time conditions.
              </p>
            </div>
          </div>
        </div>

        {/* Right: Results */}
        <div className="col-span-12 lg:col-span-7 space-y-gutter">
          {/* Error State */}
          {error && (
            <div className="card-elevation rounded-xl p-lg border-l-4 border-error bg-error-container/30">
              <div className="flex items-center gap-sm mb-sm">
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
                  className="text-error"
                >
                  <circle cx="12" cy="12" r="10" />
                  <line x1="12" y1="8" x2="12" y2="12" />
                  <line x1="12" y1="16" x2="12.01" y2="16" />
                </svg>
                <h3 className="button-text text-error">Calculation Error</h3>
              </div>
              <p className="body-sm text-on-surface-variant">{error}</p>
            </div>
          )}

          {/* Results Card */}
          {result && (
            <>
              <div className="card-elevation rounded-xl p-xl">
                <div className="flex justify-between items-center mb-lg">
                  <h3 className="title-md">Pricing Factors Breakdown</h3>
                  <span className="badge-success">ACCURATE 99.8%</span>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-lg">
                  {/* Base Rate */}
                  <div className="card-elevation rounded-lg p-md metric-accent-primary flex justify-between items-center">
                    <div>
                      <p className="label-mono text-on-surface-variant mb-base uppercase">
                        Base Rate
                      </p>
                      <p className="headline-lg">
                        ${result.base_rate.toFixed(2)}
                        <span className="body-sm text-on-surface-variant font-normal">
                          /hr
                        </span>
                      </p>
                    </div>
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      width="36"
                      height="36"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="1.5"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      className="text-primary opacity-20"
                    >
                      <rect x="2" y="5" width="20" height="14" rx="2" />
                      <line x1="2" y1="10" x2="22" y2="10" />
                    </svg>
                  </div>

                  {/* Demand Multiplier */}
                  <div className="card-elevation rounded-lg p-md metric-accent-secondary flex justify-between items-center">
                    <div>
                      <p className="label-mono text-on-surface-variant mb-base uppercase">
                        Demand Multiplier
                      </p>
                      <p className="headline-lg">
                        {result.demand_multiplier.toFixed(2)}x
                      </p>
                    </div>
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      width="36"
                      height="36"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="1.5"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      className="text-secondary opacity-20"
                    >
                      <polyline points="23 6 13.5 15.5 8.5 10.5 1 18" />
                      <polyline points="17 6 23 6 23 12" />
                    </svg>
                  </div>

                  {/* Zone Surge */}
                  <div className="card-elevation rounded-lg p-md metric-accent-tertiary flex justify-between items-center">
                    <div>
                      <p className="label-mono text-on-surface-variant mb-base uppercase">
                        Zone Surge
                      </p>
                      <p className="headline-lg">
                        +${result.zone_surge.toFixed(2)}
                      </p>
                    </div>
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      width="36"
                      height="36"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="1.5"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      className="text-tertiary opacity-20"
                    >
                      <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" />
                      <circle cx="12" cy="10" r="3" />
                    </svg>
                  </div>

                  {/* Battery Discount */}
                  <div className="card-elevation rounded-lg p-md metric-accent-secondary flex justify-between items-center">
                    <div>
                      <p className="label-mono text-on-surface-variant mb-base uppercase">
                        Battery Rebate
                      </p>
                      <p className="headline-lg text-secondary">
                        -{Math.abs(result.battery_discount).toFixed(0)}%
                      </p>
                    </div>
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      width="36"
                      height="36"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="1.5"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      className="text-secondary opacity-20"
                    >
                      <rect
                        x="1"
                        y="6"
                        width="18"
                        height="12"
                        rx="2"
                        stroke="currentColor"
                      />
                      <line x1="23" y1="10" x2="23" y2="14" />
                      <line x1="9" y1="8" x2="9" y2="16" />
                    </svg>
                  </div>
                </div>

                {/* Grand Total */}
                <div className="mt-xl pt-xl border-t border-outline-variant flex items-center justify-between">
                  <div>
                    <p className="body-md text-on-surface-variant">
                      Total Estimated Cost
                    </p>
                    <p
                      className="display-lg text-primary"
                      style={{ fontFamily: "'JetBrains Mono', monospace" }}
                    >
                      ${result.total_price.toFixed(2)}
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="label-mono text-on-surface-variant mb-xs uppercase">
                      Savings Applied
                    </p>
                    <p className="title-md text-secondary">
                      -$
                      {(
                        result.base_rate *
                          result.duration_hours *
                          (Math.abs(result.battery_discount) / 100)
                      ).toFixed(2)}
                    </p>
                  </div>
                </div>
              </div>
            </>
          )}

          {/* Empty State */}
          {!result && !error && !loading && (
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
                <path d="M18 20V10" />
                <path d="M12 20V4" />
                <path d="M6 20v-6" />
              </svg>
              <h3 className="title-md text-on-surface-variant mb-xs">
                No Calculation Yet
              </h3>
              <p className="body-sm text-on-surface-variant max-w-sm">
                Enter pricing parameters and click Calculate to see results with
                detailed cost breakdown.
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
