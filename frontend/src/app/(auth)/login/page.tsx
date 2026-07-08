"use client";

import { useState, FormEvent } from "react";
import { useRouter } from "next/navigation";
import { login } from "@/lib/api";
import { clsx } from "clsx";

export default function LoginPage() {
  const router = useRouter();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);

    const result = await login(username, password);

    if (result.data?.token) {
      router.push("/pricing");
    } else {
      setError(result.error || "Authentication failed. Please try again.");
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen flex flex-col items-center justify-center px-md py-xl relative bg-background">
      {/* Pattern overlay */}
      <div className="fixed inset-0 pattern-bg pointer-events-none" />

      <div className="w-full max-w-[440px] relative z-10">
        {/* Brand */}
        <div className="text-center mb-xl">
          <div className="flex items-center justify-center mb-xs">
            <svg
              width="44"
              height="44"
              viewBox="0 0 24 24"
              fill="none"
              stroke="#064e3b"
              strokeWidth="1.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            </svg>
          </div>
          <h1 className="headline-lg text-primary tracking-tight">Electrum</h1>
          <p className="body-sm text-on-surface-variant mt-xs">
            Secure Access to Pricing Intelligence
          </p>
        </div>

        {/* Login Card */}
        <div className="card-elevation rounded-xl p-xl">
          <form onSubmit={handleSubmit} className="space-y-lg">
            {/* Username */}
            <div className="space-y-xs">
              <label
                htmlFor="username"
                className="label-mono uppercase text-on-surface-variant block"
              >
                Username
              </label>
              <input
                id="username"
                type="text"
                required
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="admin@electrum.io"
                className="input-field w-full"
                autoComplete="username"
              />
            </div>

            {/* Password */}
            <div className="space-y-xs">
              <label
                htmlFor="password"
                className="label-mono uppercase text-on-surface-variant block"
              >
                Password
              </label>
              <input
                id="password"
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                className="input-field w-full"
                autoComplete="current-password"
              />
            </div>

            {/* Error */}
            {error && (
              <div className="p-sm rounded-lg bg-error-container text-on-error-container body-sm flex items-center gap-xs">
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
                  className="shrink-0"
                >
                  <circle cx="12" cy="12" r="10" />
                  <line x1="12" y1="8" x2="12" y2="12" />
                  <line x1="12" y1="16" x2="12.01" y2="16" />
                </svg>
                {error}
              </div>
            )}

            {/* Submit */}
            <button
              type="submit"
              disabled={loading}
              className={clsx(
                "btn-primary w-full py-md",
                loading && "opacity-70"
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
                  Authenticating...
                </>
              ) : (
                <>
                  Sign In
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
                    <line x1="5" y1="12" x2="19" y2="12" />
                    <polyline points="12 5 19 12 12 19" />
                  </svg>
                </>
              )}
            </button>
          </form>
        </div>

        {/* Status Indicator */}
        <div className="mt-lg flex items-center justify-center gap-xs text-on-surface-variant">
          <span className="w-2 h-2 rounded-full bg-secondary-fixed-dim animate-pulse" />
          <span className="label-mono">System Operational</span>
        </div>
      </div>
    </main>
  );
}
