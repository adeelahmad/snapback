// Snapback web UI enhancements. Every page works without this module; it only
// adds keyboard scrubbing, a collapsible Versions panel and in-place restore.
// Listing a file's versions needs it: file rows only carry the root and path.

function csrfToken() {
  const meta = document.querySelector('meta[name="csrf-token"]');
  return meta ? meta.getAttribute('content') : '';
}

function csrfFetch(url, options = {}) {
  const headers = new Headers(options.headers || {});
  headers.set('X-CSRF-Token', csrfToken());
  return fetch(url, { ...options, headers, credentials: 'same-origin' });
}

function initTimeline(timeline) {
  const ticks = Array.from(timeline.querySelectorAll('button[data-snapshot]'));
  ticks.forEach((tick, i) => {
    tick.addEventListener('keydown', (event) => {
      const step = event.key === 'ArrowRight' ? 1 : event.key === 'ArrowLeft' ? -1 : 0;
      const next = ticks[i + step];
      if (step !== 0 && next) {
        event.preventDefault();
        next.focus();
      }
    });
  });
}

function initVersions(panel) {
  const heading = panel.querySelector('h2');
  if (!heading) {
    return;
  }
  const toggle = document.createElement('button');
  toggle.type = 'button';
  toggle.className = 'button';
  toggle.textContent = 'Hide';
  toggle.setAttribute('aria-expanded', 'true');
  toggle.addEventListener('click', () => {
    const expanded = toggle.getAttribute('aria-expanded') === 'true';
    toggle.setAttribute('aria-expanded', String(!expanded));
    toggle.textContent = expanded ? 'Show' : 'Hide';
    panel.querySelectorAll('.group').forEach((group) => {
      group.hidden = expanded;
    });
  });
  heading.after(toggle);
}

// failure describes a non-OK API response, using the server's JSON error
// message when there is one.
async function failure(res) {
  try {
    const body = await res.json();
    if (body.error) {
      return `${res.status}: ${body.error}`;
    }
  } catch {
    // Not JSON; the status code alone has to do.
  }
  return String(res.status);
}

function hiddenInput(name, value) {
  const input = document.createElement('input');
  input.type = 'hidden';
  input.name = name;
  input.value = value;
  return input;
}

function textSpan(text) {
  const span = document.createElement('span');
  span.textContent = text;
  return span;
}

// openButton asks the server to show the folder holding file, where the
// restored copy sits next to the original.
function openButton(root, file) {
  const button = document.createElement('button');
  button.type = 'button';
  button.className = 'button has-icon icon-folder-open';
  button.textContent = 'Open in file manager';
  button.addEventListener('click', async () => {
    const body = new FormData();
    body.append('root', root);
    body.append('path', file);
    try {
      const res = await csrfFetch('/api/open', { method: 'POST', body });
      if (!res.ok) {
        button.textContent = `Open failed (${await failure(res)})`;
      }
    } catch {
      button.textContent = 'Open failed: network error';
    }
  });
  return button;
}

function initRestore(form) {
  const status = document.createElement('span');
  status.setAttribute('role', 'status');
  form.append(status);
  form.addEventListener('submit', async (event) => {
    event.preventDefault();
    status.textContent = 'Restoring…';
    try {
      const res = await csrfFetch('/api/restore', { method: 'POST', body: new FormData(form) });
      if (!res.ok) {
        status.textContent = `Restore failed (${await failure(res)}).`;
        return;
      }
      const result = await res.json();
      status.textContent = `Restored copy saved as ${result.path} `;
      status.append(openButton(form.elements.root.value, form.elements.file.value));
    } catch {
      status.textContent = 'Restore failed: network error.';
    }
  });
}

// restoreName mirrors the dated name /api/restore picks when nothing is in
// the way: "report.docx" taken at 2026-09-20 becomes "report (2026-09-20).docx".
function restoreName(file, time) {
  const base = file.slice(file.lastIndexOf('/') + 1);
  const dot = base.lastIndexOf('.');
  const ext = dot >= 0 ? base.slice(dot) : '';
  const t = new Date(time);
  const pad = (n) => String(n).padStart(2, '0');
  const day = `${t.getFullYear()}-${pad(t.getMonth() + 1)}-${pad(t.getDate())}`;
  return `${base.slice(0, base.length - ext.length)} (${day})${ext}`;
}

function versionItem(root, file, version) {
  const item = document.createElement('li');
  item.append(
    textSpan(new Date(version.time).toLocaleString()), ' ',
    textSpan(`${version.size} bytes`), ' ',
    textSpan(`modified ${new Date(version.mtime).toLocaleString()}`), ' ',
  );
  if (version.label) {
    const caution = document.createElement('p');
    caution.className = 'caution';
    caution.textContent = `${version.label} (same size and modified time)`;
    item.append(caution);
  }
  const download = document.createElement('a');
  download.className = 'button has-icon icon-download';
  download.href = `/api/download?${new URLSearchParams({ root, snapshot: version.snapshot, path: file })}`;
  download.textContent = 'Download';

  const form = document.createElement('form');
  form.method = 'post';
  form.action = '/api/restore';
  form.dataset.js = 'restore';
  const restore = document.createElement('button');
  restore.type = 'submit';
  restore.className = 'button has-icon icon-undo-2';
  restore.textContent = 'Restore copy next to original';
  form.append(
    hiddenInput('csrf_token', csrfToken()),
    hiddenInput('root', root),
    hiddenInput('snapshot', version.snapshot),
    hiddenInput('file', file),
    restore, ' ',
    textSpan(`saves as ${restoreName(file, version.time)}`),
  );
  initRestore(form);
  item.append(download, form);
  return item;
}

function renderVersions(panel, root, file, versions) {
  const list = panel.querySelector('[data-js="versions-list"]');
  if (versions.length === 0) {
    list.replaceChildren(textSpan('No snapshot holds a copy of this file.'));
    return;
  }
  const group = document.createElement('section');
  group.className = 'group';
  const toggle = panel.querySelector('button[aria-expanded]');
  group.hidden = toggle !== null && toggle.getAttribute('aria-expanded') === 'false';
  const items = document.createElement('ol');
  versions.forEach((version) => items.append(versionItem(root, file, version)));
  group.append(items);
  list.replaceChildren(group);
}

// openVersions shows every snapshot's copy of the file a row button names,
// each with Download and Restore, in the Versions panel.
async function openVersions(panel, button) {
  const { root, file } = button.dataset;
  const list = panel.querySelector('[data-js="versions-list"]');
  panel.querySelector('[data-js="versions-file"]').textContent = file;
  list.replaceChildren(textSpan('Loading…'));
  panel.hidden = false;
  try {
    const res = await csrfFetch(`/api/versions?${new URLSearchParams({ root, path: file })}`);
    if (!res.ok) {
      list.replaceChildren(textSpan(`Could not list copies (${await failure(res)}).`));
      return;
    }
    renderVersions(panel, root, file, await res.json());
  } catch {
    list.replaceChildren(textSpan('Could not list copies: network error.'));
  }
}

function initVersionsOpen(button) {
  const panel = document.querySelector('[data-js="versions"]');
  if (!panel || !panel.querySelector('[data-js="versions-list"]')) {
    return;
  }
  button.addEventListener('click', () => openVersions(panel, button));
}

document.querySelectorAll('[data-js="timeline"]').forEach(initTimeline);
document.querySelectorAll('[data-js="versions"]').forEach(initVersions);
document.querySelectorAll('[data-js="restore"]').forEach(initRestore);
document.querySelectorAll('[data-js="versions-open"]').forEach(initVersionsOpen);
