"use client";

import { useEffect, useState, FormEvent } from "react";
import {
  getConfig,
  updateConfig,
  getConfigHistory,
  PricingConfig,
  ConfigHistoryEntry,
} from "@/lib/api";
import { clsx } from "clsx";

export default function SettingsPage() {
  const [config, setConfig] = useState<PricingConfig | null>(null);
  const [history, setHistory] = useState<ConfigHistoryEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [toast, setToast] = useState<{
    type: "success" | "error";
    message: string;
  } | null>(null);

  // Editable form values
  const [basePrice, setBasePrice] = useState("14.50");
  const [surgeCap, setSurgeCap] = useState("4.5");
  const [minFare, setMinFare] = useState("12.00");
  const [corpDiscount, setCorpDiscount] = useState("25");

  useEffect(() => {
    async function load() {
      setLoading(true);
      const [configRes, historyRes] = await Promise.all([
        getConfig(),
        getConfigHistory(),
      ]);

      if (configRes.data) {
        setConfig(configRes.data);
        setBasePrice(configRes.data.base_price.toFixed(2));
        setSurgeCap(configRes.data.surge_cap.toFixed(1));
      } else if (configRes.error) {
        setError(configRes.error);
      }

      if (historyRes.data) {
        setHistory(historyRes.data);
      }

      setLoading(false);
    }
    load();
  }, []);

  async function handleSave(e: FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(null);

    const res = await updateConfig({
      base_price: parseFloat(basePrice),
      surge_cap: parseFloat(surgeCap),
    });

    if (res.data) {
      setConfig(res.data);
      showToast("success", "Configuration saved successfully.");
    } else {
      showToast("error", res.error || "Failed to save configuration.");
    }
    setSaving(false);
  }

  function showToast(type: "success" | "error", message: string) {
    setToast({ type, message });
    setTimeout(() => setToast(null), 4000);
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
    <div className="space-y-gutter relative">
      {/* Toast */}
      {toast && (
        <div
          className={clsx(
            "fixed bottom-xl right-lg z-50 px-lg py-md rounded-xl shadow-lg flex items-center gap-md toast-enter",
            toast.type === "success"
              ? "bg-secondary-container text-on-secondary-container"
              : "bg-error-container text-on-error-container"
          )}
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
            {toast.type === "success" ? (
              <>
                <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
                <polyline points="22 4 12 14.01 9 11.01" />
              </>
            ) : (
              <>
                <circle cx="12" cy="12" r="10" />
                <line x1="12" y1="8" x2="12" y2="12" />
                <line x1="12" y1="16" x2="12.01" y2="16" />
              </>
            )}
          </svg>
          <span className="button-text">{toast.message}</span>
        </div>
      )}

      <div className="grid grid-cols-12 gap-gutter">
        {/* Left Column: Profile + System */}
        <section className="col-span-12 lg:col-span-4 flex flex-col gap-gutter">
          {/* Profile */}
          <div className="card-elevation card-elevation-hover rounded-xl p-lg border-l-4 border-primary">
            <div className="flex items-center gap-md mb-lg">
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
                className="text-primary"
              >
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
                <circle cx="12" cy="7" r="4" />
              </svg>
              <h3 className="title-md">Profile</h3>
            </div>
            <form className="space-y-md">
              <div>
                <label className="label-mono text-on-surface-variant uppercase block mb-xs">
                  Full Name
                </label>
                <input
                  className="input-field"
                  type="text"
                  defaultValue="Marcus Chen"
                />
              </div>
              <div>
                <label className="label-mono text-on-surface-variant uppercase block mb-xs">
                  Email Address
                </label>
                <input
                  className="input-field"
                  type="email"
                  defaultValue="marcus.chen@electrum.io"
                />
              </div>
              <div className="pt-md border-t border-outline-variant">
                <h4 className="body-sm font-semibold mb-sm">Change Password</h4>
                <div className="space-y-sm">
                  <input
                    className="input-field"
                    type="password"
                    placeholder="Current Password"
                  />
                  <input
                    className="input-field"
                    type="password"
                    placeholder="New Password"
                  />
                </div>
              </div>
              <button
                type="button"
                className="btn-primary w-full py-md mt-md"
              >
                Update Profile
              </button>
            </form>
          </div>

          {/* Current Config JSON Viewer */}
          {config && (
            <div className="card-elevation rounded-xl p-lg border-l-4 border-primary-container">
              <div className="flex items-center gap-md mb-lg">
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
                  className="text-primary-container"
                >
                  <polyline points="16 18 22 12 16 6" />
                  <polyline points="8 6 2 12 8 18" />
                </svg>
                <h3 className="title-md">Current Config</h3>
              </div>
              <pre
                className="label-mono text-on-surface-variant bg-surface-container-low rounded-lg p-md overflow-x-auto text-xs whitespace-pre-wrap"
                style={{ fontSize: "11px", lineHeight: "1.6" }}
              >
                {JSON.stringify(config, null, 2)}
              </pre>
            </div>
          )}
        </section>

        {/* Right Column: Pricing Rules + History */}
        <section className="col-span-12 lg:col-span-8 space-y-gutter">
          {/* Global Pricing Rules */}
          <div className="card-elevation card-elevation-hover rounded-xl p-lg border-l-4 border-secondary">
            <div className="flex items-center justify-between mb-xl">
              <div className="flex items-center gap-md">
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
                  className="text-secondary"
                >
                  <path d="M18 20V10" />
                  <path d="M12 20V4" />
                  <path d="M6 20v-6" />
                </svg>
                <h3 className="title-md">Global Pricing Rules</h3>
              </div>
              <span className="badge-success">LIVE ENGINE</span>
            </div>

            <form onSubmit={handleSave}>
              <div className="grid md:grid-cols-2 gap-xl">
                <div className="space-y-lg">
                  {/* Base Rate */}
                  <div>
                    <label className="flex items-center justify-between mb-sm">
                      <span className="body-md font-semibold text-on-surface">
                        Base Rate ($/hr)
                      </span>
                      <span className="label-mono bg-surface-container-low px-xs rounded">
                        ${basePrice}
                      </span>
                    </label>
                    <input
                      type="range"
                      min="1"
                      max="50"
                      step="0.01"
                      value={basePrice}
                      onChange={(e) => setBasePrice(e.target.value)}
                    />
                    <div className="flex justify-between text-xs text-on-surface-variant mt-2">
                      <span>$1.00</span>
                      <span>$50.00</span>
                    </div>
                  </div>

                  {/* Surge Cap */}
                  <div>
                    <label className="flex items-center justify-between mb-sm">
                      <span className="body-md font-semibold text-on-surface">
                        Surge Multiplier Cap
                      </span>
                      <span className="label-mono bg-surface-container-low px-xs rounded">
                        {surgeCap}x
                      </span>
                    </label>
                    <input
                      type="range"
                      min="1"
                      max="10"
                      step="0.5"
                      value={surgeCap}
                      onChange={(e) => setSurgeCap(e.target.value)}
                    />
                    <div className="flex justify-between text-xs text-on-surface-variant mt-2">
                      <span>1.0x</span>
                      <span>10.0x</span>
                    </div>
                  </div>
                </div>

                <div className="space-y-lg">
                  {/* Minimum Fare */}
                  <div>
                    <label className="flex items-center justify-between mb-sm">
                      <span className="body-md font-semibold text-on-surface">
                        Minimum Fare Threshold
                      </span>
                      <span className="label-mono bg-surface-container-low px-xs rounded">
                        ${minFare}
                      </span>
                    </label>
                    <div className="relative">
                      <span className="absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant">
                        $
                      </span>
                      <input
                        className="input-field pl-8"
                        type="number"
                        value={minFare}
                        onChange={(e) => setMinFare(e.target.value)}
                      />
                    </div>
                  </div>

                  {/* Corporate Discount */}
                  <div>
                    <label className="flex items-center justify-between mb-sm">
                      <span className="body-md font-semibold text-on-surface">
                        Corporate Discount Cap
                      </span>
                      <span className="label-mono bg-surface-container-low px-xs rounded">
                        {corpDiscount}%
                      </span>
                    </label>
                    <div className="relative">
                      <input
                        className="input-field pr-8"
                        type="number"
                        value={corpDiscount}
                        onChange={(e) => setCorpDiscount(e.target.value)}
                      />
                      <span className="absolute right-3 top-1/2 -translate-y-1/2 text-on-surface-variant">
                        %
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              {error && (
                <div className="mt-lg p-md rounded-lg bg-error-container text-on-error-container body-sm flex items-center gap-xs">
                  {error}
                </div>
              )}

              {/* Info */}
              <div className="mt-xl p-lg bg-surface-container-low rounded-xl flex items-start gap-md">
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
                  className="text-on-surface-variant shrink-0 mt-0.5"
                >
                  <circle cx="12" cy="12" r="10" />
                  <line x1="12" y1="16" x2="12" y2="12" />
                  <line x1="12" y1="8" x2="12.01" y2="8" />
                </svg>
                <p className="body-sm text-on-surface-variant">
                  Changes to global pricing rules take effect immediately across
                  all active regions. Ensure audit logs are enabled before
                  committing large adjustments.
                </p>
              </div>

              {/* Action Buttons */}
              <div className="flex items-center justify-end gap-md pt-lg">
                <button
                  type="button"
                  className="btn-secondary px-xl py-md"
                  onClick={() => {
                    if (config) {
                      setBasePrice(config.base_price.toFixed(2));
                      setSurgeCap(config.surge_cap.toFixed(1));
                    }
                  }}
                >
                  Discard Changes
                </button>
                <button
                  type="submit"
                  disabled={saving}
                  className={clsx("btn-primary px-xl py-md", saving && "opacity-70")}
                >
                  {saving ? (
                    <>
                      <svg
                        className="animate-spin h-4 w-4"
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
                      Saving...
                    </>
                  ) : (
                    "Save All Changes"
                  )}
                </button>
              </div>
            </form>
          </div>

          {/* Config History */}
          {history.length > 0 && (
            <div className="card-elevation rounded-xl p-lg border-l-4 border-surface-tint">
              <div className="flex items-center gap-md mb-lg">
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
                  className="text-surface-tint"
                >
                  <circle cx="12" cy="12" r="10" />
                  <polyline points="12 6 12 12 16 14" />
                </svg>
                <h3 className="title-md">Version History</h3>
              </div>
              <div className="space-y-sm">
                {history.map((item, idx) => (
                  <div
                    key={idx}
                    className="flex items-center justify-between p-md rounded-lg bg-surface-container-low/50"
                  >
                    <div>
                      <span className="label-mono text-on-surface block">
                        {item.version}
                      </span>
                      <span className="body-sm text-on-surface-variant">
                        {item.summary}
                      </span>
                    </div>
                    <span className="label-mono text-on-surface-variant">
                      {new Date(item.updated_at).toLocaleDateString()}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </section>
      </div>
    </div>
  );
}
