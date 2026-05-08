// Playground app entry. Imports the published package name; Vite + tsconfig
// aliases resolve it to ../../kolavatar-client-ts/src for HMR during dev,
// so renderer edits hot-reload without a build step.
import {
  renderKolavatar,
  validateDescriptor,
  getCulturalAttribution,
  type AnimationMode,
  type Descriptor,
  type RenderOptions,
} from '@gklsndr/kolavatar-client-ts';
import { stories, type Story } from './samples.js';

// --- DOM lookups ----------------------------------------------------------

const $ = <T extends HTMLElement = HTMLElement>(sel: string): T => {
  const el = document.querySelector(sel);
  if (!el) throw new Error(`missing element: ${sel}`);
  return el as T;
};

const sidebar = $('#sidebar');
const tsSvgFrame = $('#ts-svg');
const goSvgFrame = $('#go-svg');
const goPreview = $<HTMLElement>('#go-preview');
const titleEl = $('#story-title');
const blurbEl = $('#story-blurb');
const jsonTextarea = $<HTMLTextAreaElement>('#descriptor-json');
const jsonStatus = $('#json-status');

const ctlSize = $<HTMLInputElement>('#ctl-size');
const ctlStroke = $<HTMLInputElement>('#ctl-stroke');
const ctlAnim = $<HTMLSelectElement>('#ctl-anim');
const ctlReduce = $<HTMLSelectElement>('#ctl-reduce');
const ctlDrawMs = $<HTMLInputElement>('#ctl-draw-ms');
const ctlAria = $<HTMLInputElement>('#ctl-aria');
const ctlFg = $<HTMLInputElement>('#ctl-fg');
const ctlBg = $<HTMLInputElement>('#ctl-bg');
const ctlDot = $<HTMLInputElement>('#ctl-dot');
const ctlReplay = $<HTMLButtonElement>('#ctl-replay');

// Track which color overrides the user has explicitly set, so we can leave
// them unset by default (and let the descriptor's palette through).
const colorOverrides = { fg: false, bg: false, dot: false };

// --- App state ------------------------------------------------------------

interface State {
  storyId: string;
  /** The descriptor currently being rendered — may differ from the story's
   *  default if the user edited the JSON. */
  descriptor: Descriptor;
}

const initial = stories[0]!;
const state: State = {
  storyId: initial.id,
  descriptor: structuredClone(initial.descriptor),
};

// --- Rendering ------------------------------------------------------------

function buildRenderOptions(): RenderOptions {
  const animation = ctlAnim.value as AnimationMode;
  const reduceMotion =
    ctlReduce.value === 'auto' ? 'auto' : ctlReduce.value === 'true';
  const opts: RenderOptions = {
    size: Number(ctlSize.value),
    strokeWidth: Number(ctlStroke.value),
    animation,
    reduceMotion,
    drawDurationMs: Number(ctlDrawMs.value),
    // idSuffix differentiates the inline <style> across consecutive renders so
    // CSS keyframes restart on Replay instead of resuming mid-animation.
    idSuffix: '-' + Date.now().toString(36),
  };
  const aria = ctlAria.value.trim();
  if (aria) opts.ariaLabel = aria;
  if (colorOverrides.fg || colorOverrides.bg || colorOverrides.dot) {
    opts.paletteOverride = {
      ...(colorOverrides.fg ? { fg: ctlFg.value } : {}),
      ...(colorOverrides.bg ? { bg: ctlBg.value } : {}),
      ...(colorOverrides.dot ? { dot: ctlDot.value } : {}),
    };
  }
  return opts;
}

function rerender(): void {
  try {
    validateDescriptor(state.descriptor);
  } catch (err) {
    tsSvgFrame.innerHTML =
      `<div class="muted small">descriptor invalid: ${
        (err as Error).message
      }</div>`;
    return;
  }
  try {
    const svg = renderKolavatar(state.descriptor, buildRenderOptions());
    tsSvgFrame.innerHTML = svg;
  } catch (err) {
    tsSvgFrame.innerHTML =
      `<div class="muted small">render error: ${(err as Error).message}</div>`;
  }
}

async function loadGoReference(path: string | undefined): Promise<void> {
  if (!path) {
    goPreview.hidden = true;
    goSvgFrame.innerHTML = '';
    return;
  }
  goPreview.hidden = false;
  goSvgFrame.innerHTML = '<div class="muted small">loading…</div>';
  try {
    const res = await fetch(`./go-samples/${path}`);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const text = await res.text();
    goSvgFrame.innerHTML = text;
    // Make sure the embedded SVG fits the frame at a known size.
    const svg = goSvgFrame.querySelector('svg');
    if (svg) {
      svg.setAttribute('width', String(ctlSize.value));
      svg.setAttribute('height', String(ctlSize.value));
    }
  } catch (err) {
    goSvgFrame.innerHTML =
      `<div class="muted small">missing reference (${path}): ${
        (err as Error).message
      }</div>`;
  }
}

// --- Sidebar --------------------------------------------------------------

function buildSidebar(): void {
  const groupOrder: Story['group'][] = [
    'topology',
    'symmetry',
    'edge-cases',
    'palette',
    'go-reference',
  ];
  const groupLabels: Record<Story['group'], string> = {
    topology: 'Topology',
    symmetry: 'Symmetry & complexity',
    'edge-cases': 'Edge cases',
    palette: 'Palettes',
    'go-reference': 'Go reference cross-check',
  };
  const byGroup = new Map<Story['group'], Story[]>();
  for (const s of stories) {
    if (!byGroup.has(s.group)) byGroup.set(s.group, []);
    byGroup.get(s.group)!.push(s);
  }
  const out: string[] = [];
  for (const g of groupOrder) {
    const items = byGroup.get(g);
    if (!items || items.length === 0) continue;
    out.push(`<h4>${groupLabels[g]}</h4>`);
    for (const s of items) {
      out.push(
        `<div class="story-link" data-id="${s.id}" role="button" tabindex="0">${escapeHtml(
          s.title,
        )}</div>`,
      );
    }
  }
  sidebar.innerHTML = out.join('');
  sidebar.addEventListener('click', (e) => {
    const target = e.target as HTMLElement;
    const id = target.dataset.id;
    if (id) selectStory(id);
  });
}

function setActiveLink(id: string): void {
  sidebar.querySelectorAll<HTMLElement>('.story-link').forEach((el) => {
    el.classList.toggle('active', el.dataset.id === id);
  });
}

// --- Story selection ------------------------------------------------------

function selectStory(id: string): void {
  const story = stories.find((s) => s.id === id);
  if (!story) return;
  state.storyId = story.id;
  state.descriptor = structuredClone(story.descriptor);
  setActiveLink(story.id);
  titleEl.textContent = story.title;
  blurbEl.textContent = story.blurb;
  jsonTextarea.value = JSON.stringify(state.descriptor, null, 2);
  setJsonStatus('', '');
  void loadGoReference(story.goReference);
  rerender();
}

// --- JSON editor ----------------------------------------------------------

function setJsonStatus(text: string, kind: '' | 'ok' | 'error'): void {
  jsonStatus.textContent = text;
  jsonStatus.className = 'status' + (kind ? ' ' + kind : '');
}

$('#json-apply').addEventListener('click', () => {
  let parsed: unknown;
  try {
    parsed = JSON.parse(jsonTextarea.value);
  } catch (err) {
    setJsonStatus('JSON parse error: ' + (err as Error).message, 'error');
    return;
  }
  try {
    validateDescriptor(parsed);
  } catch (err) {
    setJsonStatus('Validation: ' + (err as Error).message, 'error');
    return;
  }
  state.descriptor = parsed as Descriptor;
  setJsonStatus('Applied.', 'ok');
  rerender();
});

$('#json-reset').addEventListener('click', () => {
  selectStory(state.storyId);
});

// --- Control wiring -------------------------------------------------------

const trigger = () => rerender();
[ctlSize, ctlStroke, ctlDrawMs, ctlAria].forEach((el) =>
  el.addEventListener('input', trigger),
);
[ctlAnim, ctlReduce].forEach((el) => el.addEventListener('change', trigger));

const wireColor = (
  input: HTMLInputElement,
  clearBtn: HTMLButtonElement,
  key: 'fg' | 'bg' | 'dot',
) => {
  input.addEventListener('input', () => {
    colorOverrides[key] = true;
    trigger();
  });
  clearBtn.addEventListener('click', () => {
    colorOverrides[key] = false;
    input.value = '#000000';
    trigger();
  });
};
wireColor(ctlFg, $<HTMLButtonElement>('#ctl-fg-clear'), 'fg');
wireColor(ctlBg, $<HTMLButtonElement>('#ctl-bg-clear'), 'bg');
wireColor(ctlDot, $<HTMLButtonElement>('#ctl-dot-clear'), 'dot');

ctlReplay.addEventListener('click', () => rerender());

ctlSize.addEventListener('input', () => {
  // Keep the Go reference at the same size as our render for visual parity.
  const svg = goSvgFrame.querySelector('svg');
  if (svg) {
    svg.setAttribute('width', ctlSize.value);
    svg.setAttribute('height', ctlSize.value);
  }
});

// --- Boot -----------------------------------------------------------------

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

$('#attribution-blurb').textContent = getCulturalAttribution();
buildSidebar();
selectStory(state.storyId);
