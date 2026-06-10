// ES5-compatible API utilities — copied and adapted from shared.js
function apiCall(method, url, body, onSuccess, onUnauthorized) {
  var xhr = new XMLHttpRequest();
  xhr.onreadystatechange = function() {
    if (xhr.readyState !== 4) return;
    if (xhr.status === 401) {
      if (onUnauthorized) onUnauthorized();
      return;
    }
    var data = null;
    var contentType = xhr.getResponseHeader("content-type") || "";
    if (contentType.indexOf("application/json") !== -1) {
      try { data = JSON.parse(xhr.responseText); } catch(e) { data = { error: "Parse error" }; }
    } else {
      data = { error: xhr.responseText || "Error " + xhr.status };
    }
    onSuccess(xhr.status >= 200 && xhr.status < 300, xhr.status, data);
  };
  xhr.open(method, url, true);
  xhr.withCredentials = true;
  xhr.setRequestHeader("Content-Type", "application/json");
  // CSRF token for mutating requests
  if (method === "POST" || method === "PATCH" || method === "DELETE") {
    var csrfMatch = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]*)/);
    if (csrfMatch) {
      xhr.setRequestHeader("X-CSRF-Token", csrfMatch[1]);
    }
  }
  xhr.send(body ? JSON.stringify(body) : null);
}

function apiGet(url, onSuccess, onUnauthorized) {
  apiCall("GET", url, null, onSuccess, onUnauthorized);
}

function apiPost(url, body, onSuccess, onUnauthorized) {
  apiCall("POST", url, body, onSuccess, onUnauthorized);
}

function showLoader(show) {
  var loader = document.getElementById("loader");
  if (loader) {
    loader.classList.toggle("hidden", !show);
  }
}

function formatDate(dateStr) {
  if (!dateStr) return "";
  var d = new Date(dateStr);
  if (isNaN(d.getTime())) return "";
  function pad(n) { return n < 10 ? "0" + n : "" + n; }
  return pad(d.getDate()) + "/" + pad(d.getMonth() + 1) + "/" + d.getFullYear();
}

function playBeep(type) {
  try {
    var AudioContext = window.AudioContext || window.webkitAudioContext;
    if (!AudioContext) return;
    var ctx = new AudioContext();
    var playTone = function(freq, duration, start) {
      var osc = ctx.createOscillator();
      var gain = ctx.createGain();
      osc.connect(gain);
      gain.connect(ctx.destination);
      osc.type = "sine";
      osc.frequency.setValueAtTime(freq, ctx.currentTime + start);
      gain.gain.setValueAtTime(0.1, ctx.currentTime + start);
      osc.start(ctx.currentTime + start);
      osc.stop(ctx.currentTime + start + duration);
    };
    if (type === "success") playTone(880, 0.15, 0);
    else if (type === "warning") { playTone(660, 0.1, 0); playTone(660, 0.1, 0.15); }
    else if (type === "error") playTone(440, 0.4, 0);
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
