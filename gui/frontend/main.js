const SAVE_DEBOUNCE_MS = 50;
const DOUBLE_CLICK_RENAME_MS = 250;

const tabbar       = document.getElementById('tabbar');
const btnNewTab    = document.getElementById('btn-new-tab');
const editor       = document.getElementById('editor');
const statusSave   = document.getElementById('status-save');
const statusSaveTx = document.getElementById('status-save-text');
const statusPos    = document.getElementById('status-pos');
const statusTabs   = document.getElementById('status-tabs');
const btnMinimize  = document.getElementById('btn-minimize');
const btnMaximize  = document.getElementById('btn-maximize');
const btnClose     = document.getElementById('btn-close');

let state = { tabs: [], active_index: 0 };
let saveTimer = null;
let saving = false;

async function init() {
  try {
    state = await window.go.main.App.GetState();
    renderTabs();
    loadActiveTabIntoEditor();
  } catch (err) {
    console.error('octoNote: failed to load state', err);
    renderFallback();
  }

  document.addEventListener('keydown', onKeyDown);
}

function renderTabs() {
  tabbar.querySelectorAll('.tab:not(.tab--new)').forEach(el => el.remove());

  state.tabs.forEach((tab, i) => {
    const btn = document.createElement('button');
    btn.className = 'tab' + (i === state.active_index ? ' tab--active' : '');
    btn.setAttribute('role', 'tab');
    btn.setAttribute('aria-selected', i === state.active_index ? 'true' : 'false');
    btn.setAttribute('aria-controls', 'editor');
    btn.setAttribute('id', `tab-${tab.id}`);
    btn.setAttribute('title', `${tab.title}\nDouble-click to rename`);
    btn.dataset.index = i;

    const idxSpan = document.createElement('span');
    idxSpan.className = 'tab__index';
    idxSpan.textContent = i + 1;
    idxSpan.setAttribute('aria-hidden', 'true');

    const titleSpan = document.createElement('span');
    titleSpan.className = 'tab__title';
    titleSpan.textContent = tab.title;

    const closeBtn = document.createElement('button');
    closeBtn.className = 'tab__close';
    closeBtn.setAttribute('aria-label', `Close tab: ${tab.title}`);
    closeBtn.setAttribute('title', 'Close tab (Ctrl+W)');
    closeBtn.textContent = '×';
    closeBtn.addEventListener('click', e => {
      e.stopPropagation();
      closeTab(i);
    });

    btn.append(idxSpan, titleSpan, closeBtn);
    btn.addEventListener('click', () => switchTab(i));
    btn.addEventListener('dblclick', () => startRename(btn, titleSpan, i));

    tabbar.insertBefore(btn, btnNewTab);

    if (tab._new) {
      btn.classList.add('tab--entering');
      delete tab._new;
    }
  });

  const count = state.tabs.length;
  statusTabs.textContent = `${count} tab${count !== 1 ? 's' : ''}`;
}

function loadActiveTabIntoEditor() {
  const tab = state.tabs[state.active_index];
  if (!tab) return;
  editor.value = tab.body || '';
  try {
    const lines = editor.value.split('\n');
    let pos = 0;
    for (let l = 0; l < Math.min(tab.cursor_line || 0, lines.length - 1); l++) {
      pos += lines[l].length + 1;
    }
    editor.setSelectionRange(pos, pos);
  } catch (_) {}
  editor.focus();
}

function scheduleSave() {
  clearTimeout(saveTimer);
  setSaveStatus('saving');
  saveTimer = setTimeout(flushSave, SAVE_DEBOUNCE_MS);
}

async function flushSave() {
  const idx = state.active_index;
  const tab = state.tabs[idx];
  if (!tab) return;

  const body = editor.value;
  const cursorLine = getCursorLine();

  state.tabs[idx].body = body;
  state.tabs[idx].cursor_line = cursorLine;

  try {
    saving = true;
    await window.go.main.App.SaveTab(idx, body, cursorLine);
    setSaveStatus('saved');
  } catch (err) {
    console.error('octoNote: save failed', err);
    setSaveStatus('error');
  } finally {
    saving = false;
  }
}

async function switchTab(idx) {
  if (idx === state.active_index) return;

  clearTimeout(saveTimer);
  if (editor.value !== (state.tabs[state.active_index]?.body ?? '')) {
    await flushSave();
  }

  try {
    state = await window.go.main.App.SetActiveTab(idx);
    renderTabs();
    loadActiveTabIntoEditor();
  } catch (err) {
    console.error('octoNote: switch tab failed', err);
  }
}

async function newTab() {
  try {
    state = await window.go.main.App.NewTab();
    if (state.tabs.length > 0) {
      state.tabs[state.tabs.length - 1]._new = true;
    }
    renderTabs();
    loadActiveTabIntoEditor();
  } catch (err) {
    console.error('octoNote: new tab failed', err);
  }
}

async function closeTab(idx) {
  try {
    state = await window.go.main.App.CloseTab(idx);
    renderTabs();
    loadActiveTabIntoEditor();
  } catch (err) {
    console.error('octoNote: close tab failed', err);
  }
}

function startRename(tabBtn, titleSpan, idx) {
  if (tabBtn.querySelector('.tab-rename-input')) return;

  const originalTitle = state.tabs[idx].title;
  const input = document.createElement('input');
  input.type = 'text';
  input.className = 'tab-rename-input';
  input.value = originalTitle;
  input.maxLength = 32;
  input.setAttribute('aria-label', 'Rename tab');

  titleSpan.replaceWith(input);
  input.focus();
  input.select();

  async function commitRename() {
    const newTitle = input.value.trim() || originalTitle;
    try {
      state = await window.go.main.App.RenameTab(idx, newTitle);
      renderTabs();
    } catch (_) {
      renderTabs();
    }
  }

  input.addEventListener('keydown', e => {
    if (e.key === 'Enter') { e.preventDefault(); commitRename(); }
    if (e.key === 'Escape') { e.preventDefault(); renderTabs(); }
    e.stopPropagation();
  });

  input.addEventListener('blur', commitRename);
}

function onKeyDown(e) {
  if (document.activeElement?.classList.contains('tab-rename-input')) return;
  // Don't intercept keys when share modal is open
  if (!shareModal.hidden) return;

  const ctrl = e.ctrlKey || e.metaKey;

  if (ctrl && e.key === 'n') { e.preventDefault(); newTab(); return; }
  if (ctrl && e.key === 'w') { e.preventDefault(); closeTab(state.active_index); return; }

  if (ctrl && e.key === 'Tab' && !e.shiftKey) {
    e.preventDefault();
    const next = (state.active_index + 1) % state.tabs.length;
    switchTab(next);
    return;
  }
  if (ctrl && e.key === 'Tab' && e.shiftKey) {
    e.preventDefault();
    const prev = (state.active_index - 1 + state.tabs.length) % state.tabs.length;
    switchTab(prev);
    return;
  }
  if (ctrl && e.key === 'ArrowRight' && !e.shiftKey && document.activeElement !== editor) {
    e.preventDefault();
    switchTab((state.active_index + 1) % state.tabs.length);
    return;
  }
  if (ctrl && e.key === 'ArrowLeft' && !e.shiftKey && document.activeElement !== editor) {
    e.preventDefault();
    const prev = (state.active_index - 1 + state.tabs.length) % state.tabs.length;
    switchTab(prev);
    return;
  }

  if (ctrl && e.key >= '1' && e.key <= '9') {
    const idx = parseInt(e.key, 10) - 1;
    if (idx < state.tabs.length) {
      e.preventDefault();
      switchTab(idx);
    }
    return;
  }
}

editor.addEventListener('input', () => {
  scheduleSave();
  updateCursorStatus();
});

editor.addEventListener('keyup', updateCursorStatus);
editor.addEventListener('click', updateCursorStatus);
editor.addEventListener('selectionchange', updateCursorStatus);
document.addEventListener('selectionchange', () => {
  if (document.activeElement === editor) updateCursorStatus();
});

function getSaveTimestamp() {
  const now = new Date();
  return now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
}

function setSaveStatus(status) {
  statusSave.className = `status-indicator status-indicator--${status}`;
  if (status === 'saved') {
    statusSaveTx.textContent = `saved ${getSaveTimestamp()}`;
  } else if (status === 'saving') {
    statusSaveTx.textContent = 'saving…';
  } else {
    statusSaveTx.textContent = 'save error';
  }
}

function getCursorLine() {
  const text = editor.value.substring(0, editor.selectionStart);
  return text.split('\n').length - 1;
}

function updateCursorStatus() {
  const text = editor.value.substring(0, editor.selectionStart);
  const lines = text.split('\n');
  const line = lines.length;
  const col = lines[lines.length - 1].length + 1;
  statusPos.textContent = `Ln ${line}, Col ${col}`;
}

btnMinimize?.addEventListener('click', () => {
  window.runtime?.WindowMinimise?.();
});

btnMaximize?.addEventListener('click', () => {
  window.runtime?.WindowToggleMaximise?.();
});

btnClose?.addEventListener('click', () => {
  window.runtime?.Quit?.();
});

btnNewTab.addEventListener('click', newTab);

function renderFallback() {
  state = {
    tabs: [{ id: 'fallback', title: 'scratch', body: '', cursor_line: 0 }],
    active_index: 0,
  };
  renderTabs();
  editor.placeholder = 'Go bridge unavailable — running in offline mode.';
}

window.addEventListener('DOMContentLoaded', () => {
  let attempts = 0;
  const poll = setInterval(() => {
    if (window.go?.main?.App || attempts > 50) {
      clearInterval(poll);
      init();
    }
    attempts++;
  }, 50);
});

// ── Share modal ──────────────────────────────────────────────────────────────

const shareModal      = document.getElementById('share-modal');
const btnShare        = document.getElementById('btn-share');
const shareModalClose = document.getElementById('share-modal-close');
const shareTabHost    = document.getElementById('share-tab-host');
const shareTabGuest   = document.getElementById('share-tab-guest');
const sharePanelHost  = document.getElementById('share-panel-host');
const sharePanelGuest = document.getElementById('share-panel-guest');

// Host elements
const shareHostIdle   = document.getElementById('share-host-idle');
const shareHostActive = document.getElementById('share-host-active');
const shareHostDone   = document.getElementById('share-host-done');
const shareCodeValue  = document.getElementById('share-code-value');
const btnCopyCode     = document.getElementById('btn-copy-code');
const btnGenerateCode = document.getElementById('btn-generate-code');
const btnCancelShare  = document.getElementById('btn-cancel-share');
const shareHostStatus = document.getElementById('share-host-status');

// Guest elements
const shareGuestIdle    = document.getElementById('share-guest-idle');
const shareGuestWaiting = document.getElementById('share-guest-waiting');
const shareGuestDone    = document.getElementById('share-guest-done');
const shareGuestDoneMsg = document.getElementById('share-guest-done-msg');
const shareCodeInput    = document.getElementById('share-code-input');
const btnReceiveCode    = document.getElementById('btn-receive-code');
const btnCancelReceive  = document.getElementById('btn-cancel-receive');
const shareError        = document.getElementById('share-error');

function openShareModal() {
  resetShareModal();
  shareModal.removeAttribute('hidden');
  btnShare.classList.add('share-btn--active');
  shareModal.querySelector('.share-modal__close').focus();
}

function closeShareModal() {
  shareModal.setAttribute('hidden', '');
  btnShare.classList.remove('share-btn--active');
  // Cancel any in-flight operation
  window.go?.main?.App?.ShareCancel?.();
  editor.focus();
}

function resetShareModal() {
  shareError.setAttribute('hidden', '');
  shareError.textContent = '';
  // Host
  shareHostIdle.removeAttribute('hidden');
  shareHostActive.setAttribute('hidden', '');
  shareHostDone.setAttribute('hidden', '');
  shareCodeValue.textContent = '—';
  // Guest
  shareGuestIdle.removeAttribute('hidden');
  shareGuestWaiting.setAttribute('hidden', '');
  shareGuestDone.setAttribute('hidden', '');
  shareCodeInput.value = '';
}

function switchShareTab(tab) {
  const isHost = tab === 'host';
  shareTabHost.classList.toggle('share-tab--active', isHost);
  shareTabGuest.classList.toggle('share-tab--active', !isHost);
  shareTabHost.setAttribute('aria-selected', isHost ? 'true' : 'false');
  shareTabGuest.setAttribute('aria-selected', !isHost ? 'true' : 'false');
  sharePanelHost.toggleAttribute('hidden', !isHost);
  sharePanelGuest.toggleAttribute('hidden', isHost);
}

function showShareError(msg) {
  shareError.textContent = '⚠ ' + msg;
  shareError.removeAttribute('hidden');
}

// Open / close
btnShare.addEventListener('click', openShareModal);
shareModalClose.addEventListener('click', closeShareModal);
shareModal.querySelector('.share-modal__backdrop').addEventListener('click', closeShareModal);
document.addEventListener('keydown', e => {
  if (e.key === 'Escape' && !shareModal.hidden) closeShareModal();
});

// Tab switching
shareTabHost.addEventListener('click', () => switchShareTab('host'));
shareTabGuest.addEventListener('click', () => switchShareTab('guest'));

// ── Host: generate code ──────────────────────────────────────────────────────
btnGenerateCode.addEventListener('click', async () => {
  shareError.setAttribute('hidden', '');
  shareHostIdle.setAttribute('hidden', '');
  shareHostActive.removeAttribute('hidden');
  shareCodeValue.textContent = 'opening wormhole…';
  shareHostStatus.textContent = '⏳ Connecting to relay…';
  // Flush current editor content before sharing
  await flushSave();
  window.go.main.App.ShareSend();
});

// Copy code to clipboard
btnCopyCode.addEventListener('click', async () => {
  const code = shareCodeValue.textContent;
  if (!code || code === '—' || code === 'opening wormhole…') return;
  try {
    await navigator.clipboard.writeText(code);
    btnCopyCode.classList.add('copied');
    setTimeout(() => btnCopyCode.classList.remove('copied'), 1500);
  } catch (_) {}
});

// Cancel host share
btnCancelShare.addEventListener('click', () => {
  window.go.main.App.ShareCancel();
  resetShareModal();
});

// ── Guest: enter code & connect ─────────────────────────────────────────────
btnReceiveCode.addEventListener('click', startReceive);
shareCodeInput.addEventListener('keydown', e => {
  if (e.key === 'Enter') startReceive();
});

function startReceive() {
  const code = shareCodeInput.value.trim();
  if (!code) { showShareError('Please enter a share code.'); return; }
  shareError.setAttribute('hidden', '');
  shareGuestIdle.setAttribute('hidden', '');
  shareGuestWaiting.removeAttribute('hidden');
  window.go.main.App.ShareReceive(code);
}

btnCancelReceive.addEventListener('click', () => {
  window.go.main.App.ShareCancel();
  resetShareModal();
  switchShareTab('guest');
});

// ── Wails event listeners ────────────────────────────────────────────────────
window.addEventListener('DOMContentLoaded', () => {
  // Wait for runtime to be ready
  const waitRuntime = setInterval(() => {
    if (!window.runtime?.EventsOn) return;
    clearInterval(waitRuntime);

    // Code generated — show it
    window.runtime.EventsOn('share:code', (code) => {
      shareCodeValue.textContent = code;
      shareHostStatus.textContent = '⏳ Waiting for peer to connect…';
    });

    // Transfer complete (host side)
    window.runtime.EventsOn('share:done', () => {
      shareHostActive.setAttribute('hidden', '');
      shareHostDone.removeAttribute('hidden');
      // Auto-close after 2.5 s
      setTimeout(closeShareModal, 2500);
    });

    // Note received (guest side)
    window.runtime.EventsOn('share:received', (payload) => {
      state = payload.state;
      renderTabs();
      loadActiveTabIntoEditor();
      shareGuestWaiting.setAttribute('hidden', '');
      shareGuestDone.removeAttribute('hidden');
      shareGuestDoneMsg.textContent = `✅ Tab "${payload.title}" added to your notebook!`;
      setTimeout(closeShareModal, 2500);
    });

    // Error on either side
    window.runtime.EventsOn('share:error', (msg) => {
      resetShareModal();
      showShareError(msg);
    });
  }, 100);
});
