// src/client/shared/api.ts
async function apiCall(endpoint, options = {}, onUnauthorized) {
  try {
    const method = (options.method || "GET").toUpperCase();
    const headers = { "Content-Type": "application/json", ...options.headers };
    if (method === "POST" || method === "PATCH" || method === "DELETE") {
      const csrfCookie = document.cookie.split("; ").find((c) => c.startsWith("csrf_token="));
      if (csrfCookie) {
        headers["X-CSRF-Token"] = csrfCookie.split("=")[1];
      }
    }
    const res = await fetch(endpoint, {
      ...options,
      headers,
      credentials: "include"
    });
    if (res.status === 401) {
      onUnauthorized?.();
      return { ok: false, status: 401, data: { error: "Sessão expirada" } };
    }
    let data;
    const contentType = res.headers.get("content-type");
    if (contentType?.includes("application/json")) {
      data = await res.json();
    } else {
      const text = await res.text();
      data = { error: text || `Error ${res.status}: ${res.statusText}` };
    }
    return { ok: res.ok, status: res.status, data };
  } catch (e) {
    console.error(`API Call failed: ${endpoint}`, e);
    return {
      ok: false,
      status: 0,
      data: {
        error: e instanceof Error ? e.message : "Erro de conexão"
      }
    };
  }
}
// src/client/shared/utils.ts
function showLoader(show) {
  const loader = document.getElementById("loader");
  if (loader) {
    loader.classList.toggle("hidden", !show);
  }
}
function formatDate(dateStr) {
  if (!dateStr)
    return "";
  const d = new Date(dateStr);
  if (Number.isNaN(d.getTime()))
    return "";
  const pad = (n) => n < 10 ? `0${n}` : n;
  const day = pad(d.getDate());
  const month = pad(d.getMonth() + 1);
  const year = d.getFullYear();
  return `${day}/${month}/${year}`;
}
function playBeep(type = "success") {
  try {
    const AudioContext = window.AudioContext || window.webkitAudioContext;
    if (!AudioContext)
      return;
    const ctx = new AudioContext;
    const playTone = (freq, duration, start) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();
      osc.connect(gain);
      gain.connect(ctx.destination);
      osc.type = "sine";
      osc.frequency.setValueAtTime(freq, ctx.currentTime + start);
      gain.gain.setValueAtTime(0.1, ctx.currentTime + start);
      osc.start(ctx.currentTime + start);
      osc.stop(ctx.currentTime + start + duration);
    };
    if (type === "success")
      playTone(880, 0.15, 0);
    else if (type === "warning") {
      playTone(660, 0.1, 0);
      playTone(660, 0.1, 0.15);
    } else if (type === "error")
      playTone(440, 0.4, 0);
  } catch (_e) {}
}

function escHtml(unsafe) {
  if (unsafe == null)
    return "";
  return String(unsafe).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&#039;");
}
function sanitizeHtml(dirty) {
  return escHtml(dirty);
}
