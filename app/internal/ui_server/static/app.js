const API = '/api/routes';

function tbody() {
  return document.getElementById('routes-tbody');
}

function showError(el, msg) {
  el.textContent = msg || '';
  el.classList.toggle('hidden', !msg);
}

function escapeHtml(s) {
  const div = document.createElement('div');
  div.textContent = s;
  return div.innerHTML;
}

async function loadRoutes() {
  const res = await fetch(API);
  if (!res.ok) {
    tbody().innerHTML = '<tr><td colspan="5" class="empty">Failed to load routes</td></tr>';
    return;
  }
  const routes = await res.json();
  if (routes.length === 0) {
    tbody().innerHTML = '<tr><td colspan="5" class="empty">No routes yet. Add one to get started.</td></tr>';
    return;
  }
  tbody().innerHTML = routes.map(function (r) {
    return (
      '<tr><td>' +
      escapeHtml(r.protocol) +
      '</td><td>' +
      escapeHtml(r.method) +
      '</td><td>' +
      escapeHtml(r.source_path) +
      '</td><td>' +
      escapeHtml(r.target_path) +
      '</td><td><button type="button" onclick="editRoute(\'' +
      r.route_uuid +
      '\')">Edit</button> <button type="button" class="danger" onclick="deleteRoute(\'' +
      r.route_uuid +
      '\')">Delete</button></td></tr>'
    );
  }).join('');
}

function openCreateModal() {
  document.getElementById('create-form').reset();
  showError(document.getElementById('create-error'), '');
  document.getElementById('create-modal').classList.remove('hidden');
}

function closeCreateModal() {
  document.getElementById('create-modal').classList.add('hidden');
}

async function submitCreate(e) {
  e.preventDefault();
  const fd = new FormData(e.target);
  const errEl = document.getElementById('create-error');
  const res = await fetch(API, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      protocol: fd.get('protocol'),
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
  closeCreateModal();
  loadRoutes();
}

function closeEditModal() {
  document.getElementById('edit-modal').classList.add('hidden');
}

async function editRoute(uuid) {
  const res = await fetch(API + '/' + uuid);
  if (!res.ok) return;
  const r = await res.json();
  const form = document.getElementById('edit-form');
  form.querySelector('[name="route_uuid"]').value = r.route_uuid;
  form.querySelector('[name="protocol"]').value = r.protocol || '';
  form.querySelector('[name="method"]').value = r.method || '';
  form.querySelector('[name="source_path"]').value = r.source_path || '';
  form.querySelector('[name="target_path"]').value = r.target_path || '';
  showError(document.getElementById('edit-error'), '');
  document.getElementById('edit-modal').classList.remove('hidden');
}

async function submitEdit(e) {
  e.preventDefault();
  const fd = new FormData(e.target);
  const uuid = fd.get('route_uuid');
  const errEl = document.getElementById('edit-error');
  const res = await fetch(API + '/' + uuid, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      protocol: fd.get('protocol'),
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
  closeEditModal();
  loadRoutes();
}

async function deleteRoute(uuid) {
  if (!confirm('Delete this route?')) return;
  const res = await fetch(API + '/' + uuid, { method: 'DELETE' });
  if (res.ok) loadRoutes();
}

loadRoutes();
