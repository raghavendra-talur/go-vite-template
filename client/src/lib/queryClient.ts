import { QueryClient, QueryFunction } from "@tanstack/react-query";

const TOKEN_KEY = "app_api_token";

let resolvedApiOrigin: string | null = null;

export function getApiOrigin(): string {
  return resolvedApiOrigin ?? window.location.origin;
}

export function getStoredToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setStoredToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearStoredToken() {
  localStorage.removeItem(TOKEN_KEY);
}

function authHeaders(extra?: Record<string, string>): Record<string, string> {
  const headers: Record<string, string> = { ...extra };
  const token = getStoredToken();
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }
  return headers;
}

// Global callback for 401 responses — set by App.tsx
let onUnauthorized: (() => void) | null = null;
export function setOnUnauthorized(cb: () => void) {
  onUnauthorized = cb;
}

async function throwIfResNotOk(res: Response) {
  if (res.status === 401 && onUnauthorized) {
    onUnauthorized();
    throw new Error("Unauthorized");
  }
  if (!res.ok) {
    const text = (await res.text()) || res.statusText;
    throw new Error(`${res.status}: ${text}`);
  }
}

export async function apiRequest(
  method: string,
  url: string,
  data?: unknown | undefined,
): Promise<Response> {
  const headers = authHeaders(data ? { "Content-Type": "application/json" } : {});
  const res = await fetch(url, {
    method,
    headers,
    body: data ? JSON.stringify(data) : undefined,
  });

  if (!resolvedApiOrigin && res.url) {
    try { resolvedApiOrigin = new URL(res.url).origin; } catch {}
  }

  await throwIfResNotOk(res);
  return res;
}

export const getQueryFn: <T>() => QueryFunction<T> =
  () =>
  async ({ queryKey }) => {
    const res = await fetch(queryKey.join("/") as string, {
      headers: authHeaders(),
    });
    await throwIfResNotOk(res);
    return await res.json();
  };

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      queryFn: getQueryFn(),
      refetchInterval: false,
      refetchOnWindowFocus: false,
      staleTime: Infinity,
      retry: false,
    },
    mutations: {
      retry: false,
    },
  },
});
