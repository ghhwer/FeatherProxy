/**
 * FeatherProxy UI — App entry and UI logic.
 * Uses api.js for all backend calls.
 */

import * as api from './api.js';

// --- State (owned by UI) ---
let sourceServers = [];
let targetServers = [];
let authentications = [];
let routes = [];

// --- DOM helpers ---
function showError(el, msg) {
  el.textContent = msg || '';
  el.classList.toggle('hidden', !msg);
}

function escapeHtml(s) {
  const div = document.createElement('div');
  div.textContent = s == null ? '' : s;
  return div.innerHTML;
}

function updateStat(section, count) {
  const el = document.getElementById('stat-' + section + '-count');
  if (el) el.textContent = typeof count === 'number' ? count : '—';
}

function serverLabel(s) {
  if (s.name && s.name.trim()) return escapeHtml(s.name);
  return escapeHtml(s.host + ':' + s.port);
}

function sourceById(uuid) {
  return sourceServers.find(function (s) { return s.source_server_uuid === uuid; });
}
function targetById(uuid) {
  return targetServers.find(function (t) { return t.target_server_uuid === uuid; });
}

function protocolsCompatible(sourceProtocol, targetProtocol) {
  if (sourceProtocol === targetProtocol) return true;
  return (sourceProtocol === 'http' && targetProtocol === 'https') || (sourceProtocol === 'https' && targetProtocol === 'http');
}

// --- Source servers (UI) ---
async function loadSourceServers() {
  const tbody = document.getElementById('source-servers-tbody');
  const result = await api.getSourceServers();
  if (!result.ok) {
    tbody.innerHTML = '<tr><td colspan="5" class="empty">Failed to load source servers</td></tr>';
    updateStat('sources', '—');
    return;
  }
  sourceServers = result.data;
  updateStat('sources', sourceServers.length);
  if (sourceServers.length === 0) {
    tbody.innerHTML = '<tr><td colspan="5" class="empty">No source servers yet. Add one to get started.</td></tr>';
    return;
  }
  tbody.innerHTML = sourceServers.map(function (s) {
    return (
      '<tr><td>' + escapeHtml(s.name) +
      '</td><td>' + escapeHtml(s.protocol) +
      '</td><td>' + escapeHtml(s.host) +
      '</td><td>' + escapeHtml(String(s.port)) +
      '</td><td><button type="button" onclick="editSource(\'' + s.source_server_uuid + '\')">Edit</button> ' +
      '<button type="button" class="danger" onclick="deleteSource(\'' + s.source_server_uuid + '\')">Delete</button></td></tr>'
    );
  }).join('');
}

function openCreateSourceModal() {
  document.getElementById('create-source-form').reset();
  showError(document.getElementById('create-source-error'), '');
  document.getElementById('create-source-modal').classList.remove('hidden');
}

function closeCreateSourceModal() {
  document.getElementById('create-source-modal').classList.add('hidden');
}

async function submitCreateSource(e) {
  e.preventDefault();
  const fd = new FormData(e.target);
  const errEl = document.getElementById('create-source-error');
  const port = parseInt(fd.get('port'), 10);
  if (!port || port < 1 || port > 65535) {
    showError(errEl, 'Port must be 1–65535');
    return;
  }
  const result = await api.createSourceServer({
    name: fd.get('name') || '',
    protocol: fd.get('protocol'),
    host: fd.get('host'),
    port: port
  });
  if (!result.ok) {
    showError(errEl, result.error || 'Request failed');
    return;
  }
  closeCreateSourceModal();
  loadSourceServers();
}

function closeEditSourceModal() {
  document.getElementById('edit-source-modal').classList.add('hidden');
}

async function editSource(uuid) {
  const result = await api.getSourceServer(uuid);
  if (!result.ok) return;
  const s = result.data;
  const form = document.getElementById('edit-source-form');
  form.querySelector('[name="source_server_uuid"]').value = s.source_server_uuid;
  form.querySelector('[name="name"]').value = s.name || '';
  form.querySelector('[name="protocol"]').value = s.protocol || 'http';
  form.querySelector('[name="host"]').value = s.host || '';
  form.querySelector('[name="port"]').value = s.port || '';
  showError(document.getElementById('edit-source-error'), '');
  document.getElementById('edit-source-modal').classList.remove('hidden');
}

async function submitEditSource(e) {
  e.preventDefault();
  const fd = new FormData(e.target);
  const uuid = fd.get('source_server_uuid');
  const errEl = document.getElementById('edit-source-error');
  const port = parseInt(fd.get('port'), 10);
  if (!port || port < 1 || port > 65535) {
    showError(errEl, 'Port must be 1–65535');
    return;
  }
  const result = await api.updateSourceServer(uuid, {
    name: fd.get('name') || '',
    protocol: fd.get('protocol'),
    host: fd.get('host'),
    port: port
  });
  if (!result.ok) {
    showError(errEl, result.error || 'Request failed');
    return;
  }
  closeEditSourceModal();
  loadSourceServers();
}

async function deleteSource(uuid) {
  if (!confirm('Delete this source server?')) return;
  const result = await api.deleteSourceServer(uuid);
  if (result.ok) {
    loadSourceServers();
    loadRoutes();
  }
}

// --- Target servers (UI) ---
async function loadTargetServers() {
  const tbody = document.getElementById('target-servers-tbody');
  const result = await api.getTargetServers();
  if (!result.ok) {
    tbody.innerHTML = '<tr><td colspan="6" class="empty">Failed to load target servers</td></tr>';
    updateStat('targets', '—');
    return;
  }
  targetServers = result.data;
  updateStat('targets', targetServers.length);
  if (targetServers.length === 0) {
    tbody.innerHTML = '<tr><td colspan="6" class="empty">No target servers yet. Add one to get started.</td></tr>';
    return;
  }
  tbody.innerHTML = targetServers.map(function (t) {
    return (
      '<tr><td>' + escapeHtml(t.name) +
      '</td><td>' + escapeHtml(t.protocol) +
      '</td><td>' + escapeHtml(t.host) +
      '</td><td>' + escapeHtml(String(t.port)) +
      '</td><td>' + escapeHtml(t.base_path || '') +
      '</td><td><button type="button" onclick="editTarget(\'' + t.target_server_uuid + '\')">Edit</button> ' +
      '<button type="button" class="danger" onclick="deleteTarget(\'' + t.target_server_uuid + '\')">Delete</button></td></tr>'
    );
  }).join('');
}

function openCreateTargetModal() {
  document.getElementById('create-target-form').reset();
  showError(document.getElementById('create-target-error'), '');
  document.getElementById('create-target-modal').classList.remove('hidden');
}

function closeCreateTargetModal() {
  document.getElementById('create-target-modal').classList.add('hidden');
}

async function submitCreateTarget(e) {
  e.preventDefault();
  const fd = new FormData(e.target);
  const errEl = document.getElementById('create-target-error');
  const port = parseInt(fd.get('port'), 10);
  if (!port || port < 1 || port > 65535) {
    showError(errEl, 'Port must be 1–65535');
    return;
  }
  const result = await api.createTargetServer({
    name: fd.get('name') || '',
    protocol: fd.get('protocol'),
    host: fd.get('host'),
    port: port,
    base_path: fd.get('base_path') || ''
  });
  if (!result.ok) {
    showError(errEl, result.error || 'Request failed');
    return;
  }
  closeCreateTargetModal();
  loadTargetServers();
}

function closeEditTargetModal() {
  document.getElementById('edit-target-modal').classList.add('hidden');
}

async function editTarget(uuid) {
  const result = await api.getTargetServer(uuid);
  if (!result.ok) return;
  const t = result.data;
  const form = document.getElementById('edit-target-form');
  form.querySelector('[name="target_server_uuid"]').value = t.target_server_uuid;
  form.querySelector('[name="name"]').value = t.name || '';
  form.querySelector('[name="protocol"]').value = t.protocol || 'http';
  form.querySelector('[name="host"]').value = t.host || '';
  form.querySelector('[name="port"]').value = t.port || '';
  form.querySelector('[name="base_path"]').value = t.base_path || '';
  showError(document.getElementById('edit-target-error'), '');
  document.getElementById('edit-target-modal').classList.remove('hidden');
}

async function submitEditTarget(e) {
  e.preventDefault();
  const fd = new FormData(e.target);
  const uuid = fd.get('target_server_uuid');
  const errEl = document.getElementById('edit-target-error');
  const port = parseInt(fd.get('port'), 10);
  if (!port || port < 1 || port > 65535) {
    showError(errEl, 'Port must be 1–65535');
    return;
  }
  const result = await api.updateTargetServer(uuid, {
    name: fd.get('name') || '',
    protocol: fd.get('protocol'),
    host: fd.get('host'),
    port: port,
    base_path: fd.get('base_path') || ''
  });
  if (!result.ok) {
    showError(errEl, result.error || 'Request failed');
    return;
  }
  closeEditTargetModal();
  loadTargetServers();
}

async function deleteTarget(uuid) {
  if (!confirm('Delete this target server?')) return;
  const result = await api.deleteTargetServer(uuid);
  if (result.ok) {
    loadTargetServers();
    loadRoutes();
  }
}

// --- Authentications (UI) ---
async function loadAuthentications() {
  const tbody = document.getElementById('authentications-tbody');
  if (!tbody) return;
  const result = await api.getAuthentications();
  if (!result.ok) {
    tbody.innerHTML = '<tr><td colspan="4" class="empty">Failed to load authentications</td></tr>';
    updateStat('auths', '—');
    return;
  }
  authentications = result.data;
  updateStat('auths', authentications.length);
  if (authentications.length === 0) {
    tbody.innerHTML = '<tr><td colspan="4" class="empty">No authentications yet. Add one to attach to routes.</td></tr>';
    return;
  }
  tbody.innerHTML = authentications.map(function (a) {
    return (
      '<tr><td>' + escapeHtml(a.name) +
      '</td><td>' + escapeHtml(a.token_type || 'bearer') +
      '</td><td>' + escapeHtml(a.token_masked || '***') +
      '</td><td><button type="button" onclick="editAuth(\'' + a.authentication_uuid + '\')">Edit</button> ' +
      '<button type="button" class="danger" onclick="deleteAuth(\'' + a.authentication_uuid + '\')">Delete</button></td></tr>'
    );
  }).join('');
}

function openAuthModal(existingAuth) {
  const form = document.getElementById('auth-form');
  const uuidInput = document.getElementById('auth-form-uuid');
  const titleEl = document.getElementById('auth-modal-title');
  const tokenInput = document.getElementById('auth-form-token');
  const tokenLabel = document.getElementById('auth-form-token-label');
  const tokenHint = document.getElementById('auth-form-token-hint');
  const submitBtn = document.getElementById('auth-modal-submit');

  form.reset();
  showError(document.getElementById('auth-modal-error'), '');

  if (existingAuth) {
    uuidInput.value = existingAuth.authentication_uuid || '';
    document.getElementById('auth-form-name').value = existingAuth.name || '';
    document.getElementById('auth-form-token-type').value = existingAuth.token_type || 'bearer';
    tokenInput.value = '';
    tokenInput.removeAttribute('required');
    tokenInput.placeholder = 'New token or leave blank to keep current';
    tokenLabel.textContent = 'Token (optional)';
    tokenHint.classList.remove('hidden');
    titleEl.textContent = 'Edit authentication';
    submitBtn.textContent = 'Save';
  } else {
    uuidInput.value = '';
    tokenInput.setAttribute('required', 'required');
    tokenInput.placeholder = 'Secret token';
    tokenLabel.textContent = 'Token';
    tokenHint.classList.add('hidden');
    titleEl.textContent = 'New authentication';
    submitBtn.textContent = 'Create';
  }

  document.getElementById('auth-modal').classList.remove('hidden');
}

function closeAuthModal() {
  document.getElementById('auth-modal').classList.add('hidden');
}

async function submitAuth(e) {
  e.preventDefault();
  const fd = new FormData(e.target);
  const uuid = fd.get('authentication_uuid');
  const errEl = document.getElementById('auth-modal-error');
  const isEdit = uuid && String(uuid).trim();

  if (!isEdit) {
    const token = fd.get('token');
    if (!token || !String(token).trim()) {
      showError(errEl, 'Token is required');
      return;
    }
  }

  if (isEdit) {
    const payload = {
      name: fd.get('name') || '',
      token_type: fd.get('token_type') || 'bearer'
    };
    const token = fd.get('token');
    if (token && String(token).trim()) payload.token = token;
    const result = await api.updateAuthentication(uuid, payload);
    if (!result.ok) {
      showError(errEl, result.error || 'Request failed');
      return;
    }
  } else {
    const result = await api.createAuthentication({
      name: fd.get('name') || '',
      token_type: fd.get('token_type') || 'bearer',
      token: fd.get('token')
    });
    if (!result.ok) {
      showError(errEl, result.error || 'Request failed');
      return;
    }
  }

  closeAuthModal();
  loadAuthentications();
}

async function editAuth(uuid) {
  const result = await api.getAuthentication(uuid);
  if (!result.ok) return;
  openAuthModal(result.data);
}

async function deleteAuth(uuid) {
  if (!confirm('Delete this authentication?')) return;
  const result = await api.deleteAuthentication(uuid);
  if (result.ok) loadAuthentications();
}

// --- Routes (UI helpers + load) ---
function fillRouteSourceSelect(selectId, selectedUuid) {
  const sel = document.getElementById(selectId);
  sel.innerHTML = '<option value="">Select source server</option>' +
    sourceServers.map(function (s) {
      return '<option value="' + escapeHtml(s.source_server_uuid) + '" data-protocol="' + escapeHtml(s.protocol) + '">' +
        serverLabel(s) + ' (' + escapeHtml(s.protocol) + ')</option>';
    }).join('');
  if (selectedUuid) sel.value = selectedUuid;
}

function fillRouteTargetSelect(selectId, selectedUuid, filterByProtocol) {
  const sel = document.getElementById(selectId);
  let list = targetServers;
  if (filterByProtocol) {
    list = targetServers.filter(function (t) { return protocolsCompatible(filterByProtocol, t.protocol); });
  }
  sel.innerHTML = '<option value="">Select target server</option>' +
    list.map(function (t) {
      return '<option value="' + escapeHtml(t.target_server_uuid) + '" data-protocol="' + escapeHtml(t.protocol) + '">' +
        serverLabel(t) + ' (' + escapeHtml(t.protocol) + ')</option>';
    }).join('');
  if (selectedUuid) sel.value = selectedUuid;
}

function fillRouteAuthSelects(sourceAuthUuids, targetAuthUuid) {
  const multiSel = document.getElementById('route-auth-source-auth');
  const singleSel = document.getElementById('route-auth-target-auth');
  if (!multiSel || !singleSel) return;
  multiSel.innerHTML = authentications.map(function (a) {
    const id = a.authentication_uuid;
    const selected = sourceAuthUuids && sourceAuthUuids.indexOf(id) !== -1;
    return '<option value="' + escapeHtml(id) + '"' + (selected ? ' selected' : '') + '>' + escapeHtml(a.name || id) + '</option>';
  }).join('');
  singleSel.innerHTML = '<option value="">— None —</option>' +
    authentications.map(function (a) {
      const id = a.authentication_uuid;
      const selected = targetAuthUuid === id;
      return '<option value="' + escapeHtml(id) + '"' + (selected ? ' selected' : '') + '>' + escapeHtml(a.name || id) + '</option>';
    }).join('');
}

async function loadRoutes() {
  const tbody = document.getElementById('routes-tbody');
  const result = await api.getRoutes();
  if (!result.ok) {
    tbody.innerHTML = '<tr><td colspan="6" class="empty">Failed to load routes</td></tr>';
    updateStat('routes', '—');
    return;
  }
  routes = result.data;
  updateStat('routes', routes.length);
  if (routes.length === 0) {
    tbody.innerHTML = '<tr><td colspan="6" class="empty">No routes yet. Add source and target servers, then add a route.</td></tr>';
    return;
  }
  tbody.innerHTML = routes.map(function (r) {
    const src = sourceById(r.source_server_uuid);
    const tgt = targetById(r.target_server_uuid);
    const srcLabel = src ? serverLabel(src) : (r.source_server_uuid ? escapeHtml(r.source_server_uuid) : '—');
    const tgtLabel = tgt ? serverLabel(tgt) : (r.target_server_uuid ? escapeHtml(r.target_server_uuid) : '—');
    return (
      '<tr><td>' + escapeHtml(r.method) +
      '</td><td>' + srcLabel +
      '</td><td>' + escapeHtml(r.source_path) +
      '</td><td>' + tgtLabel +
      '</td><td>' + escapeHtml(r.target_path) +
      '</td><td><button type="button" onclick="openRouteAuthModal(\'' + r.route_uuid + '\')">Auth</button> ' +
      '<button type="button" onclick="editRoute(\'' + r.route_uuid + '\')">Edit</button> ' +
      '<button type="button" class="danger" onclick="deleteRoute(\'' + r.route_uuid + '\')">Delete</button></td></tr>'
    );
  }).join('');
}

function openCreateRouteModal() {
  fillRouteSourceSelect('create-route-source');
  fillRouteTargetSelect('create-route-target');
  document.getElementById('create-route-form').reset();
  showError(document.getElementById('create-route-error'), '');
  document.getElementById('create-route-modal').classList.remove('hidden');
}

function closeCreateRouteModal() {
  document.getElementById('create-route-modal').classList.add('hidden');
}

async function submitCreateRoute(e) {
  e.preventDefault();
  const fd = new FormData(e.target);
  const errEl = document.getElementById('create-route-error');
  const result = await api.createRoute({
    source_server_uuid: fd.get('source_server_uuid'),
    target_server_uuid: fd.get('target_server_uuid'),
    method: fd.get('method'),
    source_path: fd.get('source_path'),
    target_path: fd.get('target_path')
  });
  if (!result.ok) {
    showError(errEl, result.error || 'Request failed');
    return;
  }
  closeCreateRouteModal();
  loadRoutes();
}

function closeEditRouteModal() {
  document.getElementById('edit-route-modal').classList.add('hidden');
}

async function editRoute(uuid) {
  const result = await api.getRoute(uuid);
  if (!result.ok) return;
  const r = result.data;
  const form = document.getElementById('edit-route-form');
  form.querySelector('[name="route_uuid"]').value = r.route_uuid;
  form.querySelector('[name="method"]').value = r.method || '';
  form.querySelector('[name="source_path"]').value = r.source_path || '';
  form.querySelector('[name="target_path"]').value = r.target_path || '';
  fillRouteSourceSelect('edit-route-source', r.source_server_uuid);
  const src = sourceById(r.source_server_uuid);
  fillRouteTargetSelect('edit-route-target', r.target_server_uuid, src ? src.protocol : '');
  showError(document.getElementById('edit-route-error'), '');
  document.getElementById('edit-route-modal').classList.remove('hidden');
}

async function submitEditRoute(e) {
  e.preventDefault();
  const fd = new FormData(e.target);
  const uuid = fd.get('route_uuid');
  const errEl = document.getElementById('edit-route-error');
  const result = await api.updateRoute(uuid, {
    source_server_uuid: fd.get('source_server_uuid'),
    target_server_uuid: fd.get('target_server_uuid'),
    method: fd.get('method'),
    source_path: fd.get('source_path'),
    target_path: fd.get('target_path')
  });
  if (!result.ok) {
    showError(errEl, result.error || 'Request failed');
    return;
  }
  closeEditRouteModal();
  loadRoutes();
}

async function deleteRoute(uuid) {
  if (!confirm('Delete this route?')) return;
  const result = await api.deleteRoute(uuid);
  if (result.ok) {
    loadRoutes();
  }
}

async function openRouteAuthModal(routeUuid) {
  const [routeResult, sourceAuthResult, targetAuthResult] = await Promise.all([
    api.getRoute(routeUuid),
    api.getRouteSourceAuth(routeUuid),
    api.getRouteTargetAuth(routeUuid)
  ]);
  if (!routeResult.ok) return;
  const r = routeResult.data;
  let sourceAuthList = [];
  let targetAuthUuid = '';
  if (sourceAuthResult.ok && Array.isArray(sourceAuthResult.data)) {
    sourceAuthList = sourceAuthResult.data.map(function (x) { return x.authentication_uuid; });
  }
  if (targetAuthResult.ok && targetAuthResult.data && targetAuthResult.data.authentication_uuid) {
    targetAuthUuid = targetAuthResult.data.authentication_uuid;
  }
  document.getElementById('route-auth-route-uuid').value = routeUuid;
  document.getElementById('route-auth-modal-title').textContent = 'Route auth: ' + (r.method || '') + ' ' + (r.source_path || '');
  fillRouteAuthSelects(sourceAuthList, targetAuthUuid);
  showError(document.getElementById('route-auth-error'), '');
  document.getElementById('route-auth-modal').classList.remove('hidden');
}

function closeRouteAuthModal() {
  document.getElementById('route-auth-modal').classList.add('hidden');
}

async function submitRouteAuth(e) {
  e.preventDefault();
  const fd = new FormData(e.target);
  const uuid = fd.get('route_uuid');
  const errEl = document.getElementById('route-auth-error');
  let sourceAuthUuids = [];
  const sel = document.getElementById('route-auth-source-auth');
  if (sel) {
    for (let i = 0; i < sel.options.length; i++) {
      if (sel.options[i].selected) sourceAuthUuids.push(sel.options[i].value);
    }
  }
  let targetAuthUuid = '';
  const targetSel = document.getElementById('route-auth-target-auth');
  if (targetSel && targetSel.value) targetAuthUuid = targetSel.value;
  await Promise.all([
    api.putRouteSourceAuth(uuid, sourceAuthUuids),
    api.putRouteTargetAuth(uuid, targetAuthUuid)
  ]);
  closeRouteAuthModal();
}

// --- Event listeners (dropdowns, nav, resize) ---
document.getElementById('create-route-source').addEventListener('change', function () {
  const opt = this.options[this.selectedIndex];
  const protocol = opt ? opt.getAttribute('data-protocol') : '';
  fillRouteTargetSelect('create-route-target', '', protocol);
});
document.getElementById('edit-route-source').addEventListener('change', function () {
  const opt = this.options[this.selectedIndex];
  const protocol = opt ? opt.getAttribute('data-protocol') : '';
  fillRouteTargetSelect('edit-route-target', '', protocol);
});

function refreshAll() {
  Promise.all([loadSourceServers(), loadTargetServers(), loadAuthentications()]).then(function () {
    return loadRoutes();
  });
}

// Tab switching: show one section, hide others, update nav and title
const SECTION_IDS = ['home', 'sources', 'targets', 'auth', 'routes'];
const SECTION_TITLES = { home: 'Dashboard', sources: 'Source servers', targets: 'Target servers', auth: 'Authentications', routes: 'Routes' };

function showSection(id) {
  if (!id || SECTION_IDS.indexOf(id) === -1) return;
  const pageTitleEl = document.getElementById('page-title');
  SECTION_IDS.forEach(function (sid) {
    const section = document.getElementById(sid);
    if (section) section.classList.toggle('hidden', sid !== id);
  });
  const nav = document.querySelector('.section-nav');
  if (nav) {
    nav.querySelectorAll('.section-nav-link').forEach(function (link) {
      link.classList.toggle('active', link.getAttribute('href') === '#' + id);
    });
  }
  if (pageTitleEl && SECTION_TITLES[id]) pageTitleEl.textContent = SECTION_TITLES[id];
}

(function () {
  const nav = document.querySelector('.section-nav');
  if (!nav) return;
  nav.addEventListener('click', function (e) {
    const link = e.target.closest('a[href^="#"]');
    if (!link || !link.getAttribute('href')) return;
    const id = link.getAttribute('href').slice(1);
    if (SECTION_IDS.indexOf(id) !== -1) {
      e.preventDefault();
      showSection(id);
    }
  });
  // Home stat cards: click switches to that section
  document.querySelectorAll('.stat-card-link').forEach(function (card) {
    card.addEventListener('click', function (e) {
      const id = card.getAttribute('data-section');
      if (id && SECTION_IDS.indexOf(id) !== -1) {
        e.preventDefault();
        showSection(id);
      }
    });
  });
  // Ensure initial state: home visible, first link active
  showSection('home');
})();

// --- Globals for inline handlers in HTML ---
window.refreshAll = refreshAll;
window.openCreateSourceModal = openCreateSourceModal;
window.closeCreateSourceModal = closeCreateSourceModal;
window.submitCreateSource = submitCreateSource;
window.closeEditSourceModal = closeEditSourceModal;
window.editSource = editSource;
window.submitEditSource = submitEditSource;
window.deleteSource = deleteSource;
window.openCreateTargetModal = openCreateTargetModal;
window.closeCreateTargetModal = closeCreateTargetModal;
window.submitCreateTarget = submitCreateTarget;
window.closeEditTargetModal = closeEditTargetModal;
window.editTarget = editTarget;
window.submitEditTarget = submitEditTarget;
window.deleteTarget = deleteTarget;
window.openAuthModal = function () { openAuthModal(null); };
window.closeAuthModal = closeAuthModal;
window.submitAuth = submitAuth;
window.editAuth = editAuth;
window.deleteAuth = deleteAuth;
window.openCreateRouteModal = openCreateRouteModal;
window.closeCreateRouteModal = closeCreateRouteModal;
window.submitCreateRoute = submitCreateRoute;
window.closeEditRouteModal = closeEditRouteModal;
window.editRoute = editRoute;
window.submitEditRoute = submitEditRoute;
window.deleteRoute = deleteRoute;
window.openRouteAuthModal = openRouteAuthModal;
window.closeRouteAuthModal = closeRouteAuthModal;
window.submitRouteAuth = submitRouteAuth;

// Initial load
refreshAll();
