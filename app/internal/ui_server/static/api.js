/**
 * FeatherProxy UI — API layer.
 * All HTTP calls to the backend. No DOM access.
 * Each function returns a promise of { ok, data } or { ok: false, error }.
 */

const API_ROUTES = '/api/routes';
const API_SOURCE = '/api/source-servers';
const API_TARGET = '/api/target-servers';
const API_AUTH = '/api/authentications';

async function request(url, options = {}) {
  const res = await fetch(url, options);
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    return { ok: false, error: data.error || res.statusText };
  }
  return { ok: true, data };
}

// --- Source servers ---
export async function getSourceServers() {
  const res = await fetch(API_SOURCE);
  if (!res.ok) return { ok: false, error: (await res.json().catch(() => ({}))).error || res.statusText };
  return { ok: true, data: await res.json() };
}

export async function createSourceServer(body) {
  return request(API_SOURCE, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  });
}

export async function getSourceServer(uuid) {
  const res = await fetch(API_SOURCE + '/' + uuid);
  if (!res.ok) return { ok: false, error: (await res.json().catch(() => ({}))).error || res.statusText };
  return { ok: true, data: await res.json() };
}

export async function updateSourceServer(uuid, body) {
  return request(API_SOURCE + '/' + uuid, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  });
}

export async function deleteSourceServer(uuid) {
  const res = await fetch(API_SOURCE + '/' + uuid, { method: 'DELETE' });
  return { ok: res.ok };
}

// --- Target servers ---
export async function getTargetServers() {
  const res = await fetch(API_TARGET);
  if (!res.ok) return { ok: false, error: (await res.json().catch(() => ({}))).error || res.statusText };
  return { ok: true, data: await res.json() };
}

export async function createTargetServer(body) {
  return request(API_TARGET, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  });
}

export async function getTargetServer(uuid) {
  const res = await fetch(API_TARGET + '/' + uuid);
  if (!res.ok) return { ok: false, error: (await res.json().catch(() => ({}))).error || res.statusText };
  return { ok: true, data: await res.json() };
}

export async function updateTargetServer(uuid, body) {
  return request(API_TARGET + '/' + uuid, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  });
}

export async function deleteTargetServer(uuid) {
  const res = await fetch(API_TARGET + '/' + uuid, { method: 'DELETE' });
  return { ok: res.ok };
}

// --- Authentications ---
export async function getAuthentications() {
  const res = await fetch(API_AUTH);
  if (!res.ok) return { ok: false, error: (await res.json().catch(() => ({}))).error || res.statusText };
  return { ok: true, data: await res.json() };
}

export async function createAuthentication(body) {
  return request(API_AUTH, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  });
}

export async function getAuthentication(uuid) {
  const res = await fetch(API_AUTH + '/' + uuid);
  if (!res.ok) return { ok: false, error: (await res.json().catch(() => ({}))).error || res.statusText };
  return { ok: true, data: await res.json() };
}

export async function updateAuthentication(uuid, body) {
  return request(API_AUTH + '/' + uuid, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  });
}

export async function deleteAuthentication(uuid) {
  const res = await fetch(API_AUTH + '/' + uuid, { method: 'DELETE' });
  return { ok: res.ok };
}

// --- Routes ---
export async function getRoutes() {
  const res = await fetch(API_ROUTES);
  if (!res.ok) return { ok: false, error: (await res.json().catch(() => ({}))).error || res.statusText };
  return { ok: true, data: await res.json() };
}

export async function createRoute(body) {
  return request(API_ROUTES, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  });
}

export async function getRoute(uuid) {
  const res = await fetch(API_ROUTES + '/' + uuid);
  if (!res.ok) return { ok: false, error: (await res.json().catch(() => ({}))).error || res.statusText };
  return { ok: true, data: await res.json() };
}

export async function updateRoute(uuid, body) {
  return request(API_ROUTES + '/' + uuid, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  });
}

export async function deleteRoute(uuid) {
  const res = await fetch(API_ROUTES + '/' + uuid, { method: 'DELETE' });
  return { ok: res.ok };
}

export async function getRouteSourceAuth(uuid) {
  const res = await fetch(API_ROUTES + '/' + uuid + '/source-auth');
  if (!res.ok) return { ok: false, error: (await res.json().catch(() => ({}))).error || res.statusText };
  return { ok: true, data: await res.json() };
}

export async function getRouteTargetAuth(uuid) {
  const res = await fetch(API_ROUTES + '/' + uuid + '/target-auth');
  if (!res.ok) return { ok: false, error: (await res.json().catch(() => ({}))).error || res.statusText };
  return { ok: true, data: await res.json() };
}

export async function putRouteSourceAuth(uuid, authentication_uuids) {
  return request(API_ROUTES + '/' + uuid + '/source-auth', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ authentication_uuids })
  });
}

export async function putRouteTargetAuth(uuid, authentication_uuid) {
  return request(API_ROUTES + '/' + uuid + '/target-auth', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ authentication_uuid })
  });
}

/** POST /api/reload — trigger proxy restart so new source servers are picked up. */
export async function reloadProxies() {
  const res = await fetch('/api/reload', { method: 'POST' });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    return { ok: false, error: data.error || res.statusText };
  }
  return { ok: true, data };
}
