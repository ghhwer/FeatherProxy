const API_ROUTES = '/api/routes';
const API_SOURCE = '/api/source-servers';
const API_TARGET = '/api/target-servers';

let sourceServers = [];
let targetServers = [];
let routes = [];

function showError(el, msg) {
  el.textContent = msg || '';
  el.classList.toggle('hidden', !msg);
}

function escapeHtml(s) {
  const div = document.createElement('div');
  div.textContent = s == null ? '' : s;
  return div.innerHTML;
}

async function switchTab(tabId) {
  document.querySelectorAll('.nav-tabs .tab').forEach(function (t) {
    t.classList.toggle('active', t.getAttribute('data-tab') === tabId);
  });
  document.querySelectorAll('.tab-panel').forEach(function (p) {
    p.classList.toggle('active', p.id === 'panel-' + tabId);
  });
  if (tabId === 'source-servers') loadSourceServers();
  if (tabId === 'target-servers') loadTargetServers();
  if (tabId === 'routes') {
    await loadSourceServers();
    await loadTargetServers();
    await loadRoutes();
  }
}

function serverLabel(s) {
  if (s.name && s.name.trim()) return escapeHtml(s.name);
  return escapeHtml(s.host + ':' + s.port);
}

// --- Source servers ---
async function loadSourceServers() {
  const tbody = document.getElementById('source-servers-tbody');
  const res = await fetch(API_SOURCE);
  if (!res.ok) {
    tbody.innerHTML = '<tr><td colspan="5" class="empty">Failed to load source servers</td></tr>';
    return;
  }
  sourceServers = await res.json();
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
  const res = await fetch(API_SOURCE, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      name: fd.get('name') || '',
      protocol: fd.get('protocol'),
      host: fd.get('host'),
      port: port
    })
  });
  const data = await res.json().catch(function () { return {}; });
  if (!res.ok) {
    showError(errEl, data.error || res.statusText);
    return;
  }
  closeCreateSourceModal();
  loadSourceServers();
}

function closeEditSourceModal() {
  document.getElementById('edit-source-modal').classList.add('hidden');
}

async function editSource(uuid) {
  const res = await fetch(API_SOURCE + '/' + uuid);
  if (!res.ok) return;
  const s = await res.json();
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
  const res = await fetch(API_SOURCE + '/' + uuid, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      name: fd.get('name') || '',
      protocol: fd.get('protocol'),
      host: fd.get('host'),
      port: port
    })
  });
  const data = await res.json().catch(function () { return {}; });
  if (!res.ok) {
    showError(errEl, data.error || res.statusText);
    return;
  }
  closeEditSourceModal();
  loadSourceServers();
}

async function deleteSource(uuid) {
  if (!confirm('Delete this source server?')) return;
  const res = await fetch(API_SOURCE + '/' + uuid, { method: 'DELETE' });
  if (res.ok) loadSourceServers();
}

// --- Target servers ---
async function loadTargetServers() {
  const tbody = document.getElementById('target-servers-tbody');
  const res = await fetch(API_TARGET);
  if (!res.ok) {
    tbody.innerHTML = '<tr><td colspan="6" class="empty">Failed to load target servers</td></tr>';
    return;
  }
  targetServers = await res.json();
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
  const res = await fetch(API_TARGET, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      name: fd.get('name') || '',
      protocol: fd.get('protocol'),
      host: fd.get('host'),
      port: port,
      base_path: fd.get('base_path') || ''
    })
  });
  const data = await res.json().catch(function () { return {}; });
  if (!res.ok) {
    showError(errEl, data.error || res.statusText);
    return;
  }
  closeCreateTargetModal();
  loadTargetServers();
}

function closeEditTargetModal() {
  document.getElementById('edit-target-modal').classList.add('hidden');
}

async function editTarget(uuid) {
  const res = await fetch(API_TARGET + '/' + uuid);
  if (!res.ok) return;
  const t = await res.json();
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
  const res = await fetch(API_TARGET + '/' + uuid, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      name: fd.get('name') || '',
      protocol: fd.get('protocol'),
      host: fd.get('host'),
      port: port,
      base_path: fd.get('base_path') || ''
    })
  });
  const data = await res.json().catch(function () { return {}; });
  if (!res.ok) {
    showError(errEl, data.error || res.statusText);
    return;
  }
  closeEditTargetModal();
  loadTargetServers();
}

async function deleteTarget(uuid) {
  if (!confirm('Delete this target server?')) return;
  const res = await fetch(API_TARGET + '/' + uuid, { method: 'DELETE' });
  if (res.ok) loadTargetServers();
}

// --- Routes ---
function sourceById(uuid) {
  return sourceServers.find(function (s) { return s.source_server_uuid === uuid; });
}
function targetById(uuid) {
  return targetServers.find(function (t) { return t.target_server_uuid === uuid; });
}

function fillRouteSourceSelect(selectId, selectedUuid) {
  const sel = document.getElementById(selectId);
  sel.innerHTML = '<option value="">Select source server</option>' +
    sourceServers.map(function (s) {
      return '<option value="' + escapeHtml(s.source_server_uuid) + '" data-protocol="' + escapeHtml(s.protocol) + '">' +
        serverLabel(s) + ' (' + escapeHtml(s.protocol) + ')</option>';
    }).join('');
  if (selectedUuid) sel.value = selectedUuid;
}

function protocolsCompatible(sourceProtocol, targetProtocol) {
  if (sourceProtocol === targetProtocol) return true;
  return (sourceProtocol === 'http' && targetProtocol === 'https') || (sourceProtocol === 'https' && targetProtocol === 'http');
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

async function loadRoutes() {
  const tbody = document.getElementById('routes-tbody');
  const res = await fetch(API_ROUTES);
  if (!res.ok) {
    tbody.innerHTML = '<tr><td colspan="6" class="empty">Failed to load routes</td></tr>';
    return;
  }
  routes = await res.json();
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
      '</td><td><button type="button" onclick="editRoute(\'' + r.route_uuid + '\')">Edit</button> ' +
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
  const res = await fetch(API_ROUTES, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      source_server_uuid: fd.get('source_server_uuid'),
      target_server_uuid: fd.get('target_server_uuid'),
      method: fd.get('method'),
      source_path: fd.get('source_path'),
      target_path: fd.get('target_path')
    })
  });
  const data = await res.json().catch(function () { return {}; });
  if (!res.ok) {
    showError(errEl, data.error || res.statusText);
    return;
  }
  closeCreateRouteModal();
  loadRoutes();
}

function closeEditRouteModal() {
  document.getElementById('edit-route-modal').classList.add('hidden');
}

async function editRoute(uuid) {
  const res = await fetch(API_ROUTES + '/' + uuid);
  if (!res.ok) return;
  const r = await res.json();
  const form = document.getElementById('edit-route-form');
  form.querySelector('[name="route_uuid"]').value = r.route_uuid;
  form.querySelector('[name="method"]').value = r.method || '';
  form.querySelector('[name="source_path"]').value = r.source_path || '';
  form.querySelector('[name="target_path"]').value = r.target_path || '';
  fillRouteSourceSelect('edit-route-source', r.source_server_uuid);
  var src = sourceById(r.source_server_uuid);
  fillRouteTargetSelect('edit-route-target', r.target_server_uuid, src ? src.protocol : '');
  showError(document.getElementById('edit-route-error'), '');
  document.getElementById('edit-route-modal').classList.remove('hidden');
}

async function submitEditRoute(e) {
  e.preventDefault();
  const fd = new FormData(e.target);
  const uuid = fd.get('route_uuid');
  const errEl = document.getElementById('edit-route-error');
  const res = await fetch(API_ROUTES + '/' + uuid, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      source_server_uuid: fd.get('source_server_uuid'),
      target_server_uuid: fd.get('target_server_uuid'),
      method: fd.get('method'),
      source_path: fd.get('source_path'),
      target_path: fd.get('target_path')
    })
  });
  const data = await res.json().catch(function () { return {}; });
  if (!res.ok) {
    showError(errEl, data.error || res.statusText);
    return;
  }
  closeEditRouteModal();
  loadRoutes();
}

async function deleteRoute(uuid) {
  if (!confirm('Delete this route?')) return;
  const res = await fetch(API_ROUTES + '/' + uuid, { method: 'DELETE' });
  if (res.ok) loadRoutes();
}

// Filter target server dropdown by source server protocol (same-protocol rule)
document.getElementById('create-route-source').addEventListener('change', function () {
  var opt = this.options[this.selectedIndex];
  var protocol = opt ? opt.getAttribute('data-protocol') : '';
  fillRouteTargetSelect('create-route-target', '', protocol);
});
document.getElementById('edit-route-source').addEventListener('change', function () {
  var opt = this.options[this.selectedIndex];
  var protocol = opt ? opt.getAttribute('data-protocol') : '';
  fillRouteTargetSelect('edit-route-target', '', protocol);
});

// Initial load: show source servers tab and load its data
loadSourceServers();
