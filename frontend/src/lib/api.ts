/* ===== Electrum DPE API Client ===== */

// Use relative URL — nginx proxies /api/v1/* to backend
const DEFAULT_BASE_URL = "/api/v1";

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
        errorMsg = errData.message || errData.error || errorMsg;
      } catch {
        // Could not parse error JSON
      }
      return { data: null, error: errorMsg, status };
    }

    // Handle empty responses (204)
    const text = await res.text();
    if (!text) {
      return { data: null as T, error: null, status };
    }

    const data = JSON.parse(text);
    return { data, error: null, status };
  } catch (err) {
    const message = err instanceof Error ? err.message : "Network error";
    return { data: null, error: message, status: 0 };
  }
}

/* ===== Auth ===== */

export interface LoginResult {
  token: string;
  expires_at: number;
  username: string;
  role: string;
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

export interface PricingBreakdown {
  base_rate_per_hour: number;
  demand_multiplier: number;
  zone_surge_factor: number;
  battery_discount_factor: number;
}

export interface PricingResult {
  vehicle_id: string;
  zone: string;
  duration_hours: number;
  total_price: number;
  currency: string;
  breakdown: PricingBreakdown;
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

export interface AuditFactors {
  base_rate_per_hour: number;
  demand_multiplier: number;
  zone_surge_factor: number;
  battery_discount_factor: number;
}

export interface AuditLogEntry {
  id: string;
  timestamp: string;
  vehicle_id: string;
  zone: string;
  duration_hours: number;
  final_price: number;
  factors_applied: AuditFactors;
  config_version: number;
  signature: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export async function getAuditLogs(
  page: number = 1,
  page_size: number = 20,
  filters?: { vehicle_id?: string; zone?: string }
): Promise<ApiResponse<PaginatedResponse<AuditLogEntry>>> {
  const query = new URLSearchParams({
    page: String(page),
    page_size: String(page_size),
  });
  if (filters?.vehicle_id) query.set("vehicle_id", filters.vehicle_id);
  if (filters?.zone) query.set("zone", filters.zone);
  return request<PaginatedResponse<AuditLogEntry>>(
    "GET",
    `/admin/pricing/audit?${query.toString()}`
  );
}

/* ===== Config ===== */

export interface ConfigResult {
  base_price_per_hour: number;
  currency: string;
  surge_cap_multiplier: number;
  demand_multipliers: unknown;
  zone_surge_config: unknown;
  battery_discount_tiers: unknown;
  version: number;
  created_at: string;
  updated_at: string;
}

export interface ConfigHistoryEntry {
  version: number;
  base_price_per_hour: number;
  currency: string;
  surge_cap_multiplier: number;
  demand_multipliers: unknown;
  zone_surge_config: unknown;
  battery_discount_tiers: unknown;
  created_at: string;
}

export async function getConfig(): Promise<ApiResponse<ConfigResult>> {
  return request<ConfigResult>("GET", "/admin/config");
}

export async function updateConfig(
  data: Record<string, unknown>
): Promise<ApiResponse<ConfigResult>> {
  return request<ConfigResult>("PUT", "/admin/config", data);
}

export async function getConfigHistory(
  page: number = 1,
  page_size: number = 10
): Promise<ApiResponse<PaginatedResponse<ConfigHistoryEntry>>> {
  return request<PaginatedResponse<ConfigHistoryEntry>>(
    "GET",
    `/admin/config/history?page=${page}&page_size=${page_size}`
  );
}

/* ===== Zones ===== */

export interface ZoneInfo {
  name: string;
  utilization: number;
}

export async function getZones(): Promise<ApiResponse<ZoneInfo[]>> {
  return request<ZoneInfo[]>("GET", "/zones");
}
