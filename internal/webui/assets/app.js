// Snapback web UI enhancements. Every page works without this module; it only
// adds keyboard scrubbing, a collapsible Versions panel and in-place restore.

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

function initRestore(form) {
  const status = document.createElement('span');
  status.setAttribute('role', 'status');
  form.append(status);
  form.addEventListener('submit', async (event) => {
    event.preventDefault();
    status.textContent = 'Restoring…';
    try {
      const res = await csrfFetch('/api/restore', { method: 'POST', body: new FormData(form) });
      status.textContent = res.ok ? 'Restored copy saved.' : `Restore failed (${res.status}).`;
    } catch {
      status.textContent = 'Restore failed: network error.';
    }
  });
}

function reindexNames(container) {
  const counts = new Map();
  container.querySelectorAll('[name]').forEach((el) => {
    const name = el.getAttribute('name');
    const match = /\[(\d+)\]$/.exec(name);
    if (!match) {
      return;
    }
    const stem = name.slice(0, name.length - match[0].length);
    const next = counts.has(stem) ? counts.get(stem) + 1 : 0;
    counts.set(stem, next);
    el.setAttribute('name', stem + '[' + next + ']');
    const oldID = el.getAttribute('id');
    const idIndex = oldID ? /\[(\d+)\]$/.exec(oldID) : null;
    if (idIndex) {
      const newID = oldID.slice(0, oldID.length - idIndex[0].length) + '[' + next + ']';
      container.querySelectorAll('label[for="' + oldID + '"]').forEach((label) => {
        label.setAttribute('for', newID);
      });
      el.setAttribute('id', newID);
    }
  });
}

function initRows(root) {
  root.addEventListener('click', (event) => {
    const add = event.target.closest('[data-js="row-add"]');
    if (add) {
      const rows = root.querySelectorAll('.field__row');
      const last = rows[rows.length - 1];
      const row = last ? last.cloneNode(true) : document.createElement('div');
      row.className = 'field__row';
      row.querySelectorAll('input').forEach((input) => {
        input.value = '';
        input.removeAttribute('value');
      });
      add.before(row);
      reindexNames(root);
      return;
    }
    const button = event.target.closest('[data-js="row-remove"]');
    const row = button ? button.closest('.field__row') : null;
    if (row) {
      row.remove();
      reindexNames(root);
    }
  });
}

function initChips(input) {
  const field = input.closest('.field--chips');
  const list = field ? field.querySelector('.field__chips') : null;
  if (!list) {
    return;
  }
  input.addEventListener('keydown', (event) => {
    const value = input.value.trim();
    if (event.key !== 'Enter' || value === '') {
      return;
    }
    event.preventDefault();
    const item = document.createElement('li');
    item.className = 'chip';
    const hidden = document.createElement('input');
    hidden.type = 'hidden';
    hidden.setAttribute('name', input.getAttribute('name'));
    hidden.setAttribute('value', value);
    const text = document.createElement('span');
    text.className = 'chip__text';
    text.textContent = value;
    const drop = document.createElement('button');
    drop.type = 'button';
    drop.className = 'button-secondary chip__remove';
    drop.setAttribute('data-js', 'chip-remove');
    drop.textContent = 'Remove ' + value;
    item.append(hidden, text, drop);
    list.append(item);
    input.value = '';
    reindexNames(field);
  });
  field.addEventListener('click', (event) => {
    const button = event.target.closest('[data-js="chip-remove"]');
    const item = button ? button.closest('.chip') : null;
    if (item) {
      item.remove();
      reindexNames(field);
    }
  });
}

function initDaemonStatus(pill) {
  const refresh = async () => {
    try {
      const res = await csrfFetch('/api/daemon');
      if (!res.ok) {
        return;
      }
      const state = await res.json();
      pill.textContent = state.running ? 'running' : 'stopped';
    } catch (err) {
      // Leave the server-rendered text in place when the poll fails.
    }
  };
  setInterval(refresh, 5000);
}

document.querySelectorAll('[data-js="timeline"]').forEach(initTimeline);
document.querySelectorAll('[data-js="versions"]').forEach(initVersions);
document.querySelectorAll('[data-js="restore"]').forEach(initRestore);
document.querySelectorAll('[data-js="rows"]').forEach(initRows);
document.querySelectorAll('[data-js="chips"]').forEach(initChips);
document.querySelectorAll('[data-js="daemon-status"]').forEach(initDaemonStatus);
