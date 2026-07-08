/* ===== Electrum DPE API Client ===== */

const DEFAULT_BASE_URL = "http://localhost:8080/api/v1";

function getBaseUrl(): string {
  if (typeof window !== "undefined") {
    return localStorage.getItem("api_base_url") || DEFAULT_BASE_URL;
  }
  return DEFAULT_BASE_URL;
}

function getToken(): string | null {
  if (typeof window !== "undefined") {
    return localStorage.getItem("auth_token");
  }
  return null;
}

function setToken(token: string): void {
  if (typeof window !== "undefined") {
    localStorage.setItem("auth_token", token);
  }
}

function clearToken(): void {
  if (typeof window !== "undefined") {
    localStorage.removeItem("auth_token");
  }
}

function isAuthenticated(): boolean {
  return getToken() !== null;
}

interface ApiResponse<T> {
  data: T | null;
  error: string | null;
  status: number;
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown
): Promise<ApiResponse<T>> {
  const baseUrl = getBaseUrl();
  const token = getToken();
  const url = `${baseUrl}${path}`;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  try {
    const res = await fetch(url, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    });

    const status = res.status;

    if (!res.ok) {
      let errorMsg = `HTTP ${status}`;
      try {
        const errData = await res.json();
        errorMsg = errData.error || errData.message || errorMsg;
      } catch {
        // Could not parse error JSON
      }
      return { data: null, error: errorMsg, status };
    }

    const data = await res.json();
    return { data, error: null, status };
  } catch (err) {
    const message = err instanceof Error ? err.message : "Network error";
    return { data: null, error: message, status: 0 };
  }
}

/* ===== Auth ===== */

export interface LoginPayload {
  username: string;
  password: string;
}

export interface LoginResult {
  token: string;
  user?: {
    id: string;
    username: string;
    role: string;
  };
}

export async function login(
  username: string,
  password: string
): Promise<ApiResponse<LoginResult>> {
  const result = await request<LoginResult>("POST", "/auth/login", {
    username,
    password,
  });
  if (result.data?.token) {
    setToken(result.data.token);
  }
  return result;
}

export function logout(): void {
  clearToken();
  if (typeof window !== "undefined") {
    window.location.href = "/login";
  }
}

export { isAuthenticated, getToken, clearToken, setToken };

/* ===== Pricing ===== */

export interface PricingParams {
  vehicle_id: string;
  zone: string;
  duration_hours: number;
}

export interface PricingResult {
  vehicle_id: string;
  zone: string;
  duration_hours: number;
  total_price: number;
  base_rate: number;
  demand_multiplier: number;
  zone_surge: number;
  battery_discount: number;
  currency: string;
  calculated_at: string;
}

export async function calculatePricing(
  params: PricingParams
): Promise<ApiResponse<PricingResult>> {
  const query = new URLSearchParams({
    vehicle_id: params.vehicle_id,
    zone: params.zone,
    duration_hours: String(params.duration_hours),
  }).toString();
  return request<PricingResult>("GET", `/pricing?${query}`);
}

/* ===== Audit Logs ===== */

export interface AuditLogEntry {
  id: string;
  timestamp: string;
  vehicle_id: string;
  zone: string;
  duration_hours: number;
  final_price: number;
  factors: Record<string, number>;
  user?: string;
}

export interface AuditLogResponse {
  entries: AuditLogEntry[];
  total: number;
  page: number;
  page_size: number;
}

export async function getAuditLogs(
  page: number = 1,
  page_size: number = 20,
  filters?: { vehicle_id?: string; zone?: string }
): Promise<ApiResponse<AuditLogResponse>> {
  const query = new URLSearchParams({
    page: String(page),
    page_size: String(page_size),
  });
  if (filters?.vehicle_id) query.set("vehicle_id", filters.vehicle_id);
  if (filters?.zone) query.set("zone", filters.zone);
  return request<AuditLogResponse>("GET", `/pricing/audit?${query.toString()}`);
}

/* ===== Config ===== */

export interface PricingConfig {
  version: string;
  base_price: number;
  surge_cap: number;
  demand_multipliers: Record<string, number>;
  zone_surge: Record<string, number>;
  battery_discounts: Record<string, number>;
  updated_at: string;
}

export interface ConfigHistoryEntry {
  version: string;
  updated_at: string;
  summary: string;
}

export async function getConfig(): Promise<ApiResponse<PricingConfig>> {
  return request<PricingConfig>("GET", "/admin/config");
}

export async function updateConfig(
  data: Partial<PricingConfig>
): Promise<ApiResponse<PricingConfig>> {
  return request<PricingConfig>("PUT", "/admin/config", data);
}

export async function getConfigHistory(): Promise<
  ApiResponse<ConfigHistoryEntry[]>
> {
  return request<ConfigHistoryEntry[]>("GET", "/admin/config/history");
}

/* ===== Zones ===== */

export interface ZoneInfo {
  id: string;
  name: string;
  code: string;
  utilization: number;
  status: "active" | "inactive" | "peak";
  multiplier: number;
  demand: "low" | "medium" | "high";
  surge_threshold: number;
}

export interface ZonesSummary {
  zones: ZoneInfo[];
  total_active: number;
  avg_multiplier: number;
  high_demand_peak: number;
}

export async function getZones(): Promise<ApiResponse<ZonesSummary>> {
  return request<ZonesSummary>("GET", "/zones");
}
