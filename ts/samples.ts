// Curated descriptor catalogue for the playground.
//
// Hand-crafted because the Go package's gallery only ships server-rendered
// SVGs — descriptors aren't archived. Each fixture targets a specific
// rendering behaviour the TS port needs to demonstrate. Tile values are
// hand-laid out to be parity-clean (each cell has an even count of active
// edges), so the strand tracer treats them as honest closed loops.
//
// Bit layout (matches kolavatar-go): N=1, E=2, S=4, W=8.
// Two-edge patterns: 3=NE, 5=NS, 6=ES, 9=NW, 10=EW, 12=SW. X cell: 15.

import type { Descriptor } from '@gklsndr/kolavatar-ts';
import { SPEC_VERSION } from '@gklsndr/kolavatar-ts';

interface Story {
  id: string;
  title: string;
  /** What this fixture demonstrates — shown in the playground sidebar. */
  blurb: string;
  /** Optional category for sidebar grouping. */
  group: 'edge-cases' | 'symmetry' | 'topology' | 'palette' | 'go-reference';
  /** Optional path under /go-samples/ to show as a server-side reference. */
  goReference?: string;
  descriptor: Descriptor;
}

const baseGenerated = {
  package: 'kolavatar',
  version: SPEC_VERSION,
  hash_func: 'blake3',
  cultural_attribution: 'Sahapedia, Siromoney 1974, Gopalan 2024',
};

const indigoDawn = {
  name: 'indigo-dawn',
  colors: ['#1a237e', '#fafafa', '#5c6bc0', '#e8eaf6'],
};

const margazhi = {
  name: 'margazhi-classic',
  colors: ['#5b0f00', '#fdfdfa', '#7a3e3e', '#b08a4f'],
};

const banana = {
  name: 'banana-leaf',
  colors: ['#1b5e20', '#fffde7', '#558b2f', '#fbc02d'],
};

const monoBw = {
  name: 'test-bw',
  colors: ['#ffffff', '#000000', '#888888', '#444444'],
};

function descriptor(overrides: Partial<Descriptor>): Descriptor {
  return {
    spec_version: SPEC_VERSION,
    seed_hash: 'fixturefixture00',
    tier: 4,
    symmetry_group: '4mm_d',
    grid: 7,
    template: '1R',
    tiles: [],
    edge_states: '00',
    palette: indigoDawn,
    generated_by: baseGenerated,
    ...overrides,
  };
}

// ---------------------------------------------------------------------------
// Hand-crafted tile arrays
// ---------------------------------------------------------------------------

// 3×3: pure dot grid. No paths, just dots — verifies the renderer handles the
// "newly initialised" state gracefully.
const tilesEmpty3 = [
  0, 0, 0,
  0, 0, 0,
  0, 0, 0,
];

// 3×3: 8-cell perimeter loop around the centre dot. The simplest closed
// kolam; a single strand traces all 8 cells.
const tilesPerimeter3 = [
  6, 10, 12,
  5,  0,  5,
  3, 10,  9,
];

// 5×5: 16-cell perimeter — same idea, larger.
const tilesPerimeter5 = [
  6, 10, 10, 10, 12,
  5,  0,  0,  0,  5,
  5,  0,  0,  0,  5,
  5,  0,  0,  0,  5,
  3, 10, 10, 10,  9,
];

// 5×5: perimeter frame + interior X-cell mosaic. Demonstrates how the strand
// tracer splits X cells (curl flag = (r+c)%2==0).
const tilesPerimeterPlusXFill5 = [
  6, 10, 10, 10, 12,
  5, 15, 15, 15,  5,
  5, 15, 15, 15,  5,
  5, 15, 15, 15,  5,
  3, 10, 10, 10,  9,
];

// 7×7: nested rings — outer perimeter + a 3-cell inner perimeter centred at
// (3,3). Two disjoint closed strands; both 4mm_d-symmetric.
const tilesNestedRings7 = [
  6, 10, 10, 10, 10, 10, 12,
  5,  0,  0,  0,  0,  0,  5,
  5,  0,  6, 10, 12,  0,  5,
  5,  0,  5,  0,  5,  0,  5,
  5,  0,  3, 10,  9,  0,  5,
  5,  0,  0,  0,  0,  0,  5,
  3, 10, 10, 10, 10, 10,  9,
];

// 7×7: four disjoint corner-petals (each is a 4-cell closed loop). Demonstrates
// multi-strand rendering and 4mm_d symmetry.
const tilesQuadrantPetals7 = [
  6, 12,  0, 0, 0,  6, 12,
  3,  9,  0, 0, 0,  3,  9,
  0,  0,  0, 0, 0,  0,  0,
  0,  0,  0, 0, 0,  0,  0,
  0,  0,  0, 0, 0,  0,  0,
  6, 12,  0, 0, 0,  6, 12,
  3,  9,  0, 0, 0,  3,  9,
];

// 9×9: a "labyrinth" — outer perimeter, a middle ring, and a tiny centre loop.
// Three concentric strands.
const tilesLabyrinth9 = [
  6, 10, 10, 10, 10, 10, 10, 10, 12,
  5,  0,  0,  0,  0,  0,  0,  0,  5,
  5,  0,  6, 10, 10, 10, 12,  0,  5,
  5,  0,  5,  0,  0,  0,  5,  0,  5,
  5,  0,  5,  0, 15,  0,  5,  0,  5,
  5,  0,  5,  0,  0,  0,  5,  0,  5,
  5,  0,  3, 10, 10, 10,  9,  0,  5,
  5,  0,  0,  0,  0,  0,  0,  0,  5,
  3, 10, 10, 10, 10, 10, 10, 10,  9,
];

// 11×11: nested rings + inner X-mosaic. Tier-7 grade complexity.
function buildOrnament11(): number[] {
  const N = 11;
  const t = new Array(N * N).fill(0);
  const setRow = (r: number, vals: number[]) => {
    for (let c = 0; c < N; c++) t[r * N + c] = vals[c]!;
  };
  // Outer perimeter
  setRow(0, [6, 10, 10, 10, 10, 10, 10, 10, 10, 10, 12]);
  for (let r = 1; r < N - 1; r++) {
    t[r * N] = 5;
    t[r * N + N - 1] = 5;
  }
  setRow(N - 1, [3, 10, 10, 10, 10, 10, 10, 10, 10, 10, 9]);
  // Inner ring at row 2..8
  setRow(2, [5, 0, 6, 10, 10, 10, 10, 10, 12, 0, 5]);
  for (let r = 3; r < N - 3; r++) {
    t[r * N + 2] = 5;
    t[r * N + N - 3] = 5;
  }
  setRow(N - 3, [5, 0, 3, 10, 10, 10, 10, 10, 9, 0, 5]);
  // Innermost: X-cell mosaic in the 4..6 box.
  for (let r = 4; r <= 6; r++) {
    for (let c = 4; c <= 6; c++) {
      t[r * N + c] = 15;
    }
  }
  // Frame the X mosaic so its outer edges close.
  t[3 * N + 4] = 5; t[3 * N + 5] = 5; t[3 * N + 6] = 5;
  t[7 * N + 4] = 5; t[7 * N + 5] = 5; t[7 * N + 6] = 5;
  t[4 * N + 3] = 10; t[5 * N + 3] = 10; t[6 * N + 3] = 10;
  t[4 * N + 7] = 10; t[5 * N + 7] = 10; t[6 * N + 7] = 10;
  // Corner-pin elbows where the inner-ring frame meets the X mosaic.
  t[3 * N + 3] = 6; t[3 * N + 7] = 12;
  t[7 * N + 3] = 3; t[7 * N + 7] = 9;
  return t;
}
const tilesOrnament11 = buildOrnament11();

// 5×5: deliberately asymmetric — for visualising the symmetry_group="1"
// branch where every edge bit is independent.
const tilesAsymmetric5 = [
  6, 12,  0,  6, 12,
  3,  9,  0,  3,  9,
  0,  0,  6, 10, 12,
  6, 12,  5,  0,  5,
  3,  9,  3, 10,  9,
];

// ---------------------------------------------------------------------------
// Stories
// ---------------------------------------------------------------------------

export const stories: Story[] = [
  {
    id: 'empty-3',
    title: '3×3 — empty (dots only)',
    blurb: 'Verifies the renderer paints a clean dot grid when no strands exist.',
    group: 'edge-cases',
    descriptor: descriptor({
      grid: 3,
      tiles: tilesEmpty3,
      symmetry_group: '4mm_d',
      tier: 0,
      seed_hash: 'empty-grid-fixture',
    }),
  },
  {
    id: 'perimeter-3',
    title: '3×3 — single perimeter loop',
    blurb: 'Smallest interesting kolam: one closed strand around the centre dot.',
    group: 'topology',
    descriptor: descriptor({
      grid: 3,
      tiles: tilesPerimeter3,
      symmetry_group: '4mm_d',
      tier: 1,
      seed_hash: 'perim3-fixture0000',
    }),
  },
  {
    id: 'perimeter-5',
    title: '5×5 — perimeter loop',
    blurb: 'Same closed-loop idea on a 5×5 grid; tests longer arc chaining.',
    group: 'topology',
    descriptor: descriptor({
      grid: 5,
      tiles: tilesPerimeter5,
      symmetry_group: '4mm_d',
      tier: 2,
      palette: margazhi,
      seed_hash: 'perim5-margazhi000',
    }),
  },
  {
    id: 'x-mosaic-5',
    title: '5×5 — X-mosaic interior',
    blurb:
      'Interior cells set to X (pattern 15). The (r+c)%2 curl flag splits each X into two segments, so each X cell shares a dot but joins different strands.',
    group: 'topology',
    descriptor: descriptor({
      grid: 5,
      tiles: tilesPerimeterPlusXFill5,
      symmetry_group: '4mm_d',
      tier: 3,
      palette: banana,
      seed_hash: 'xmosaic5-fixture00',
    }),
  },
  {
    id: 'nested-rings-7',
    title: '7×7 — nested rings',
    blurb: 'Two disjoint closed strands. Demonstrates per-strand path emission.',
    group: 'topology',
    descriptor: descriptor({
      grid: 7,
      tiles: tilesNestedRings7,
      symmetry_group: '4mm_d',
      tier: 4,
      seed_hash: 'nestedrings7-fix00',
    }),
  },
  {
    id: 'quadrant-petals-7',
    title: '7×7 — quadrant petals',
    blurb: 'Four small disjoint closed loops, one per corner. Tests multi-strand draw-animation staggering.',
    group: 'symmetry',
    descriptor: descriptor({
      grid: 7,
      tiles: tilesQuadrantPetals7,
      symmetry_group: '4mm_d',
      tier: 4,
      palette: margazhi,
      seed_hash: 'petals7-margazhi00',
    }),
  },
  {
    id: 'labyrinth-9',
    title: '9×9 — labyrinth',
    blurb: 'Three concentric strands of decreasing size, with a single X cell at the centre.',
    group: 'topology',
    descriptor: descriptor({
      grid: 9,
      tiles: tilesLabyrinth9,
      symmetry_group: '4mm_d',
      tier: 5,
      seed_hash: 'labyrinth9-fixture',
    }),
  },
  {
    id: 'ornament-11',
    title: '11×11 — ornament',
    blurb: 'High-tier complexity: outer + inner perimeters with an X-mosaic core. ~Tier 7 visual mass.',
    group: 'symmetry',
    descriptor: descriptor({
      grid: 11,
      tiles: tilesOrnament11,
      symmetry_group: '4mm_d',
      tier: 7,
      palette: indigoDawn,
      seed_hash: 'ornament11-fixture',
    }),
  },
  {
    id: 'asymmetric-5',
    title: '5×5 — asymmetric',
    blurb: 'Deliberately broken symmetry — what symmetry_group="1" output looks like.',
    group: 'symmetry',
    descriptor: descriptor({
      grid: 5,
      tiles: tilesAsymmetric5,
      symmetry_group: '1',
      tier: 2,
      palette: monoBw,
      seed_hash: 'asym5-bw-fixture00',
    }),
  },
  {
    id: 'go-ref-grid-7-4mm_d',
    title: 'Go reference — 7×7 4mm_d',
    blurb: 'The Go renderer\'s reference SVG (server-side). The TS-rendered fixture beside it should be visually equivalent. Use this to cross-check the strand tracer.',
    group: 'go-reference',
    goReference: 'grid-7-4mm_d.svg',
    descriptor: descriptor({
      grid: 7,
      tiles: tilesNestedRings7,
      symmetry_group: '4mm_d',
      tier: 4,
      seed_hash: 'go-ref-7-4mmd0000',
    }),
  },
  {
    id: 'go-ref-grid-11-2mm',
    title: 'Go reference — 11×11 2mm',
    blurb: 'Cross-check against the Go gallery thumbnail at samples/grid-11-2mm.svg.',
    group: 'go-reference',
    goReference: 'grid-11-2mm.svg',
    descriptor: descriptor({
      grid: 11,
      tiles: tilesOrnament11,
      symmetry_group: '2mm',
      tier: 6,
      seed_hash: 'go-ref-11-2mm0000',
    }),
  },
];

export const palettes = {
  'indigo-dawn': indigoDawn,
  'margazhi-classic': margazhi,
  'banana-leaf': banana,
  'test-bw': monoBw,
};

export type { Story };
