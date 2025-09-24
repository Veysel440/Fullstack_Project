const API: string = (import.meta.env.VITE_API_URL as string) ?? "http://localhost:8080";

type ReqInit = RequestInit & { _retry?: boolean };

export const AUTH_LOGOUT = "auth:logout";

export const tokenStore = {
    get access() { return localStorage.getItem("access_token") ?? ""; },
    set access(v: string) { localStorage.setItem("access_token", v); },
    get refresh() { return localStorage.getItem("refresh_token") ?? ""; },
    set refresh(v: string) { localStorage.setItem("refresh_token", v); },
    clear() { localStorage.removeItem("access_token"); localStorage.removeItem("refresh_token"); }
};

export async function apiFetch(path: string, init: ReqInit = {}): Promise<Response> {
    const { _retry, headers: h, ...rest } = init;
    const headers = new Headers(h ?? {});
    if (!headers.has("Authorization") && tokenStore.access) {
        headers.set("Authorization", `Bearer ${tokenStore.access}`);
    }
    let res = await fetch(API + path, { ...rest, headers });

    if (res.status === 401 && !_retry) {
        const ok = await refreshTokens();
        if (ok) {
            const headers2 = new Headers(h ?? {});
            headers2.set("Authorization", `Bearer ${tokenStore.access}`);
            res = await fetch(API + path, { ...rest, headers: headers2 });
        }
    }
    return res;
}

export async function refreshTokens(): Promise<boolean> {
    if (!tokenStore.refresh) return false;
    const r = await fetch(API + "/auth/refresh", {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-Refresh-Token": tokenStore.refresh }
    });
    if (!r.ok) { tokenStore.clear(); window.dispatchEvent(new Event(AUTH_LOGOUT)); return false; }
    const j = await r.json().catch(() => null as any);
    if (!j?.access_token) { tokenStore.clear(); window.dispatchEvent(new Event(AUTH_LOGOUT)); return false; }
    tokenStore.access = j.access_token;
    tokenStore.refresh = j.refresh_token ?? tokenStore.refresh;
    return true;
}

export async function json<T = unknown>(path: string, init?: ReqInit): Promise<T> {
    const r = await apiFetch(path, init);
    if (!r.ok) throw await safeErr(r);
    return r.json() as Promise<T>;
}

async function safeErr(r: Response) {
    try { return await r.json(); } catch { return { error: `HTTP ${r.status}` }; }
}
