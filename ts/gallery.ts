// Gallery page entry. Standalone from main.ts — fetches the static
// catalogue at ./gallery.json (Vite serves ts/public/* at the bundle root)
// and renders all entries as a wall of TS-rendered SVGs. Click any tile
// to deep-link into the playground with that descriptor pre-loaded.

import {
  renderKolavatar,
  type Descriptor,
} from '@gklsndr/kolavatar-ts';

const $ = <T extends HTMLElement = HTMLElement>(sel: string): T => {
  const el = document.querySelector(sel);
  if (!el) throw new Error(`missing element: ${sel}`);
  return el as T;
};

const galleryWall = $('#gallery-wall');
const gallerySym = $<HTMLSelectElement>('#gallery-filter-sym');
const galleryStatus = $('#gallery-status');

interface GalleryEntry {
  id: string;
  title: string;
  seed: string;
  score: number;
  grid: number;
  symmetry: string;
  descriptor: Descriptor;
}

let entries: GalleryEntry[] | null = null;

async function load(): Promise<void> {
  galleryStatus.textContent = 'loading…';
  galleryStatus.className = 'status';
  try {
    const res = await fetch('./gallery.json');
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    entries = (await res.json()) as GalleryEntry[];
    galleryStatus.textContent = `${entries.length} designs`;
    galleryStatus.className = 'status ok';
    renderWall();
  } catch (err) {
    galleryStatus.textContent = `failed to load: ${(err as Error).message}`;
    galleryStatus.className = 'status error';
  }
}

function renderWall(): void {
  if (!entries) return;
  const filterSym = gallerySym.value;
  const filtered = filterSym
    ? entries.filter((e) => e.symmetry === filterSym)
    : entries;
  const out: string[] = [];
  for (const e of filtered) {
    let svg: string;
    try {
      svg = renderKolavatar(e.descriptor, {
        size: 160,
        strokeWidth: 3,
        animation: 'none',
        reduceMotion: true,
      });
    } catch (err) {
      svg = `<div class="muted small">render error: ${
        escapeHtml((err as Error).message)
      }</div>`;
    }
    out.push(
      `<figure class="gallery-tile" data-id="${escapeHtml(e.id)}" title="${escapeHtml(e.id)}">` +
        `<div class="gallery-thumb">${svg}</div>` +
        `<figcaption>` +
          `<span class="gallery-tile-title">${escapeHtml(e.title)}</span>` +
          `<span class="muted small">${escapeHtml(e.symmetry)}</span>` +
        `</figcaption>` +
      `</figure>`,
    );
  }
  galleryWall.innerHTML = out.join('');
}

gallerySym.addEventListener('change', renderWall);

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

void load();
