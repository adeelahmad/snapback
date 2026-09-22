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

document.querySelectorAll('[data-js="timeline"]').forEach(initTimeline);
document.querySelectorAll('[data-js="versions"]').forEach(initVersions);
document.querySelectorAll('[data-js="restore"]').forEach(initRestore);
