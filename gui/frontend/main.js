/**
 * nottaker — GUI frontend JavaScript
 *
 * Communicates with Go via Wails JS bridge (window.go.main.App.*).
 * Implements: tab rendering, keyboard shortcuts, debounced auto-save,
 * cursor tracking, inline tab rename, window controls.
 */

// ── Wails bridge import ────────────────────────────────────────────────────
// Wails v2 injects window.go at runtime; we reference it after DOMContentLoaded.
// We also import the Wails runtime for window control functions.

// ── Constants ─────────────────────────────────────────────────────────────
const SAVE_DEBOUNCE_MS = 50;   // ms after last keystroke before saving
const DOUBLE_CLICK_RENAME_MS = 250;  // threshold for rename on double-click

// ── DOM refs ──────────────────────────────────────────────────────────────
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

// ── App state (mirrors Go core.State) ────────────────────────────────────
let state = { tabs: [], active_index: 0 };
let saveTimer = null;       // debounce handle
let saving = false;         // true while a save is in flight

// ── Initialisation ────────────────────────────────────────────────────────
async function init() {
  try {
    state = await window.go.main.App.GetState();
    renderTabs();
    loadActiveTabIntoEditor();
  } catch (err) {
    console.error('nottaker: failed to load state', err);
    // Render an empty fallback so UI is not broken.
    renderFallback();
  }

  // Wire up keyboard shortcuts globally.
  document.addEventListener('keydown', onKeyDown);
}

// ── Tab Rendering ─────────────────────────────────────────────────────────
function renderTabs() {
  // Remove all existing tab buttons (not the new-tab button).
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

    // Index badge
    const idxSpan = document.createElement('span');
    idxSpan.className = 'tab__index';
    idxSpan.textContent = i + 1;
    idxSpan.setAttribute('aria-hidden', 'true');

    // Title
    const titleSpan = document.createElement('span');
    titleSpan.className = 'tab__title';
    titleSpan.textContent = tab.title;

    // Close button
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

    // Insert before the new-tab button.
    tabbar.insertBefore(btn, btnNewTab);

    // Animate entry for freshly-created tabs.
    if (tab._new) {
      btn.classList.add('tab--entering');
      delete tab._new;
    }
  });

  // Update status bar.
  const count = state.tabs.length;
  statusTabs.textContent = `${count} tab${count !== 1 ? 's' : ''}`;
}

// ── Editor Load/Save ──────────────────────────────────────────────────────
function loadActiveTabIntoEditor() {
  const tab = state.tabs[state.active_index];
  if (!tab) return;
  editor.value = tab.body || '';
  // Restore cursor.
  try {
    const lines = editor.value.split('\n');
    let pos = 0;
    for (let l = 0; l < Math.min(tab.cursor_line || 0, lines.length - 1); l++) {
      pos += lines[l].length + 1;
    }
    editor.setSelectionRange(pos, pos);
  } catch (_) { /* ignore */ }
  editor.focus();
}

/** Called on every input event (debounced). */
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

  // Update local state immediately so we don't lose it on re-render.
  state.tabs[idx].body = body;
  state.tabs[idx].cursor_line = cursorLine;

  try {
    saving = true;
    await window.go.main.App.SaveTab(idx, body, cursorLine);
    setSaveStatus('saved');
  } catch (err) {
    console.error('nottaker: save failed', err);
    setSaveStatus('error');
  } finally {
    saving = false;
  }
}

// ── Tab Operations ────────────────────────────────────────────────────────
async function switchTab(idx) {
  if (idx === state.active_index) return;

  // Flush any pending save for current tab first.
  clearTimeout(saveTimer);
  if (editor.value !== (state.tabs[state.active_index]?.body ?? '')) {
    await flushSave();
  }

  try {
    state = await window.go.main.App.SetActiveTab(idx);
    renderTabs();
    loadActiveTabIntoEditor();
  } catch (err) {
    console.error('nottaker: switch tab failed', err);
  }
}

async function newTab() {
  try {
    state = await window.go.main.App.NewTab();
    // Mark last tab as new for animation.
    if (state.tabs.length > 0) {
      state.tabs[state.tabs.length - 1]._new = true;
    }
    renderTabs();
    loadActiveTabIntoEditor();
  } catch (err) {
    console.error('nottaker: new tab failed', err);
  }
}

async function closeTab(idx) {
  try {
    state = await window.go.main.App.CloseTab(idx);
    renderTabs();
    loadActiveTabIntoEditor();
  } catch (err) {
    console.error('nottaker: close tab failed', err);
  }
}

// ── Inline Tab Rename ─────────────────────────────────────────────────────
function startRename(tabBtn, titleSpan, idx) {
  if (tabBtn.querySelector('.tab-rename-input')) return; // already renaming

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
      renderTabs(); // fallback: re-render with old title
    }
  }

  input.addEventListener('keydown', e => {
    if (e.key === 'Enter') { e.preventDefault(); commitRename(); }
    if (e.key === 'Escape') { e.preventDefault(); renderTabs(); }
    e.stopPropagation(); // don't trigger global shortcuts
  });

  input.addEventListener('blur', commitRename);
}

// ── Keyboard Shortcuts ────────────────────────────────────────────────────
function onKeyDown(e) {
  // Don't intercept if user is renaming a tab.
  if (document.activeElement?.classList.contains('tab-rename-input')) return;

  const ctrl = e.ctrlKey || e.metaKey;

  if (ctrl && e.key === 'n') { e.preventDefault(); newTab(); return; }
  if (ctrl && e.key === 'w') { e.preventDefault(); closeTab(state.active_index); return; }

  // Next tab: Ctrl+Tab
  if (ctrl && e.key === 'Tab' && !e.shiftKey) {
    e.preventDefault();
    const next = (state.active_index + 1) % state.tabs.length;
    switchTab(next);
    return;
  }
  // Prev tab: Ctrl+Shift+Tab
  if (ctrl && e.key === 'Tab' && e.shiftKey) {
    e.preventDefault();
    const prev = (state.active_index - 1 + state.tabs.length) % state.tabs.length;
    switchTab(prev);
    return;
  }
  // Ctrl+Right / Ctrl+Left
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

  // Number shortcuts: Ctrl+1 … Ctrl+9
  if (ctrl && e.key >= '1' && e.key <= '9') {
    const idx = parseInt(e.key, 10) - 1;
    if (idx < state.tabs.length) {
      e.preventDefault();
      switchTab(idx);
    }
    return;
  }
}

// ── Editor Events ─────────────────────────────────────────────────────────
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

// ── Status helpers ────────────────────────────────────────────────────────
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

// ── Window Controls (Wails runtime) ───────────────────────────────────────
btnMinimize?.addEventListener('click', () => {
  window.runtime?.WindowMinimise?.();
});

btnMaximize?.addEventListener('click', () => {
  window.runtime?.WindowToggleMaximise?.();
});

btnClose?.addEventListener('click', () => {
  window.runtime?.Quit?.();
});

// ── New tab button ────────────────────────────────────────────────────────
btnNewTab.addEventListener('click', newTab);

// ── Fallback when Go bridge isn't available ───────────────────────────────
function renderFallback() {
  state = {
    tabs: [{ id: 'fallback', title: 'scratch', body: '', cursor_line: 0 }],
    active_index: 0,
  };
  renderTabs();
  editor.placeholder = 'Go bridge unavailable — running in offline mode.';
}

// ── Boot ──────────────────────────────────────────────────────────────────
// Wait for Wails runtime to be ready before calling Go methods.
window.addEventListener('DOMContentLoaded', () => {
  // Wails sets window.go asynchronously; poll briefly.
  let attempts = 0;
  const poll = setInterval(() => {
    if (window.go?.main?.App || attempts > 50) {
      clearInterval(poll);
      init();
    }
    attempts++;
  }, 50);
});
