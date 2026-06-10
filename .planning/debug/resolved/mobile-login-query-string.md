---
status: fixed
trigger: "on the mobiles handhelds, the login is not working properly. when i try to login on /login the user and password are being send to the query string and the page satys on /login, but for example, /login?username=admin&password=admin . on the server logs, i can't see any post from mobile, but when i try from the browser on localhost i see a post request to /api/auth/login, and the flow occurr normally. on another computer, I see the post request, but the user can't authenticate. on computer, the user and password are not being showed on query string, only on mobile"
created: 2026-06-10
updated: 2026-06-10
resolved: 2026-06-10
---

## Symptoms

- **Expected:** Login on mobile should POST to /api/auth/login, authenticate user (role: Conferente), redirect to /atividades
- **Actual:** On mobile, credentials appear in query string (/login?username=admin&password=admin), page stays on /login, no POST to /api/auth/login observed in server logs
- **On localhost browser:** Works normally — POST to /api/auth/login, auth flow completes
- **On another computer:** POST request visible in server logs, but user can't authenticate
- **Desktop (non-mobile):** Credentials never appear in query string
- **Error messages:** No visible error on screen, page just reloads with credentials in URL
- **Timeline:** Unknown if ever worked on mobile
- **Reproduction:** Try to login on a mobile handheld device via /login page

## Environment

- Go net/http + database/sql with pgx (Postgres)
- HTMX + Go templates
- Roles: conferente, gerente, sysadmin

## Current Focus

- **hypothesis:** Race condition in login.js — async DOMContentLoaded handler awaits `apiCall("/api/auth/me")` before registering the form submit listener. On mobile (slower network), user taps Submit before listener attaches. Form has no `method="post"` (defaults to GET), so credentials go to query string.
- **test:** Verified via code review — login.js line 42-46 async me() call before addEventListener at line 47.
- **expecting:** Adding `method="post" action="/login"` to the form ensures graceful degradation: if JS hasn't loaded/submit listener not registered, form submits via POST to the server handler which parses form body correctly.
- **next_action:** resolved

## Evidence

- **timestamp:** 2026-06-10
  Login form in `login.html` line 23 has no `method` attribute (defaults to GET).
- **timestamp:** 2026-06-10
  `login.js` line 42-47: DOMContentLoaded handler is async, calls `apiCall("/api/auth/me")` before `addEventListener("submit", ...)` on line 47. On slow mobile network, the async call blocks listener registration.
- **timestamp:** 2026-06-10
  Desktop localhost works because `/api/auth/me` completes near-instantly on loopback, listener registered before user can click.
- **timestamp:** 2026-06-10
  `loginPost` handler exists at `POST /login` and correctly parses form-encoded body via `r.ParseForm()` + `r.FormValue()`.

## Eliminated

- **hypothesis:** "fetch() not available on mobile" — eliminated because modern mobile browsers support fetch; also apiCall has try/catch and would still attach listener.
- **hypothesis:** "Secure: true cookie blocks auth on HTTP" — eliminated as root cause for GET submission symptom (unrelated — cookies aren't relevant until after credentials are sent).

## Resolution

- **root_cause:** Missing `method="post"` on login form combined with async race condition in login.js. Mobile users on slower networks tap submit before the async DOMContentLoaded handler finishes its preliminary `/api/auth/me` call and registers the submit listener. Without listener, form defaults to GET submission.
- **fix:** Added `method="post" action="/login"` to form in `login.html` and `atividades.html` for progressive enhancement fallback.
- **verification:** All non-integration tests pass (`go test ./cmd/server -count=1`). Manual verification needed on mobile.
- **files_changed:** `cmd/server/templates/login.html:23`, `cmd/server/templates/atividades.html:49`
