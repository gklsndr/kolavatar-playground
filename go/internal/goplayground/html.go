//go:build kolavatardev

package goplayground

// playgroundHTML is the single-file frontend: a form with controls and an
// SVG preview that re-fetches ./render on every change. Uses *relative*
// URLs so the same HTML works whether the page is served at / or at any
// nested prefix (e.g. /go/).
var playgroundHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>kolavatar — playground</title>
<style>
:root {
  --bg: #fafafa;
  --fg: #222;
  --accent: #2563eb;
  --panel-bg: #fff;
  --border: #ddd;
  --muted: #777;
}
body { font-family: system-ui, sans-serif; margin: 0; background: var(--bg); color: var(--fg); }
header { padding: 14px 24px; border-bottom: 1px solid var(--border); background: var(--panel-bg); }
h1 { margin: 0; font-size: 18px; }
.subtitle { color: var(--muted); font-size: 13px; margin-top: 2px; }
main { display: grid; grid-template-columns: 360px 1fr; gap: 24px; padding: 24px; }
.controls { background: var(--panel-bg); padding: 16px; border-radius: 8px; border: 1px solid var(--border); }
.preview { display: flex; flex-direction: column; align-items: center; }
.preview-svg { background: var(--panel-bg); padding: 12px; border-radius: 8px; border: 1px solid var(--border); width: 480px; height: 480px; display: flex; align-items: center; justify-content: center; }
.preview-svg svg { width: 100%; height: 100%; }
.meta { margin-top: 12px; font-size: 13px; color: var(--muted); display: flex; gap: 16px; flex-wrap: wrap; max-width: 480px; justify-content: center; }
.meta span { background: var(--panel-bg); padding: 4px 10px; border-radius: 4px; border: 1px solid var(--border); }
.field { margin-bottom: 14px; }
.field label { display: block; font-size: 12px; font-weight: 600; color: var(--fg); margin-bottom: 4px; }
.field .hint { font-size: 11px; color: var(--muted); margin-top: 2px; }
.row { display: flex; gap: 8px; align-items: center; }
input[type=text] { flex: 1; padding: 6px 8px; border: 1px solid var(--border); border-radius: 4px; font-family: monospace; font-size: 13px; }
input[type=range] { flex: 1; }
.range-val { font-family: monospace; font-size: 12px; color: var(--muted); width: 36px; text-align: right; }
button { padding: 6px 10px; border: 1px solid var(--border); background: var(--panel-bg); border-radius: 4px; cursor: pointer; font-size: 13px; }
button:hover { background: #f0f0f0; }
button.primary { background: var(--accent); color: #fff; border-color: var(--accent); }
button.primary:hover { background: #1d4ed8; }
select { padding: 6px 8px; border: 1px solid var(--border); border-radius: 4px; font-size: 13px; width: 100%; }
.field-row { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
code { font-family: ui-monospace, monospace; font-size: 12px; background: #eee; padding: 1px 4px; border-radius: 3px; }
.share { margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--border); }
.share input { width: 100%; padding: 4px 6px; font-family: monospace; font-size: 11px; border: 1px solid var(--border); border-radius: 4px; }
</style>
</head>
<body>
<header>
  <h1>kolavatar playground</h1>
  <div class="subtitle">Adjust parameters to see how seed and score interact with coverage, strand length, and motif pool.</div>
</header>
<main>
<form class="controls" id="form" onsubmit="return false;">

<div class="field">
  <label for="seed">Seed</label>
  <div class="row">
    <input type="text" id="seed" value="asha@example.com">
    <button type="button" onclick="randomizeSeed()">🎲 shuffle</button>
  </div>
  <div class="hint">Hash of seed picks symmetry group + palette. Same seed → same identity.</div>
</div>

<div class="field">
  <label for="score">Score (complexity dial)</label>
  <div class="row">
    <input type="range" id="score" min="0" max="1" step="0.01" value="0.5" oninput="document.getElementById('scoreVal').textContent = this.value">
    <span class="range-val" id="scoreVal">0.5</span>
  </div>
  <div class="hint">0.0 = sparse compact · 1.0 = dense helix. Drives coverage, strand length, motif pool size.</div>
</div>

<div class="field-row">
  <div class="field">
    <label for="grid">Grid size</label>
    <select id="grid">
      <option value="5">5×5</option>
      <option value="7">7×7</option>
      <option value="9">9×9</option>
      <option value="11" selected>11×11</option>
      <option value="13">13×13</option>
      <option value="15">15×15</option>
    </select>
  </div>
  <div class="field">
    <label for="sym">Symmetry</label>
    <select id="sym">
      <option value="auto" selected>auto (seed-driven)</option>
      <option value="1">1 (none)</option>
      <option value="m">m (h-mirror)</option>
      <option value="m_d">m_d (v-mirror)</option>
      <option value="2">2 (rot 180°)</option>
      <option value="2mm">2mm</option>
      <option value="2m_dm_d">2m_dm_d</option>
      <option value="4">4 (rot 90°)</option>
      <option value="4mm_d">4mm_d (D4)</option>
    </select>
  </div>
</div>

<div class="field">
  <label for="coverage">Coverage override</label>
  <div class="row">
    <input type="range" id="coverage" min="-0.05" max="1" step="0.05" value="-0.05" oninput="updateCoverageLabel()">
    <span class="range-val" id="coverageVal">auto</span>
  </div>
  <div class="hint">"auto" lets score decide. Drag right to force higher fill. Range: −0.05 (auto) → 1.0.</div>
</div>

<div class="field">
  <label for="strand_length">Strand length override</label>
  <select id="strand_length">
    <option value="auto" selected>auto (score-driven)</option>
    <option value="0">0 — compact (no merge)</option>
    <option value="1">1 — balanced (gentle merge)</option>
    <option value="2">2 — long (aggressive merge)</option>
    <option value="3">3 — helix (LERW + aggressive merge)</option>
  </select>
</div>

<div class="field-row">
  <div class="field">
    <label for="min_score">Min-score floor</label>
    <div class="row">
      <input type="range" id="min_score" min="0" max="1" step="0.05" value="0" oninput="document.getElementById('minScoreVal').textContent = parseFloat(this.value).toFixed(2)">
      <span class="range-val" id="minScoreVal">0.00</span>
    </div>
    <div class="hint">Clamps score from below. 0 = no floor.</div>
  </div>
  <div class="field">
    <label for="score_bands">Score bands</label>
    <select id="score_bands">
      <option value="">7 (default)</option>
      <option value="1">1 (no banding)</option>
      <option value="3">3</option>
      <option value="5">5</option>
      <option value="7">7</option>
      <option value="10">10</option>
      <option value="14">14</option>
    </select>
    <div class="hint">Same seed at different bands → different layouts.</div>
  </div>
</div>

<div class="field">
  <label for="palette">Palette</label>
  <select id="palette"></select>
  <div class="hint">"auto" = seed-derived. "B/W test" = white strands on black for visual inspection.</div>
</div>

<div class="row" style="gap: 8px;">
  <button type="button" class="primary" onclick="render()">↻ regenerate</button>
  <button type="button" onclick="downloadSVG()">⬇ download SVG</button>
</div>

<div class="share">
  <label for="share-url">Permalink (copy to share):</label>
  <input id="share-url" readonly>
</div>

</form>

<div class="preview">
  <div class="preview-svg" id="preview">
    <span style="color: var(--muted);">Loading…</span>
  </div>
  <div class="meta" id="meta"></div>
</div>

</main>

<script>
const PALETTES = ["auto", "bw", ` + paletteList() + `];
const palSel = document.getElementById("palette");
PALETTES.forEach(name => {
  const opt = document.createElement("option");
  opt.value = name;
  opt.textContent = name === "auto" ? "auto (seed-derived)" : (name === "bw" ? "B/W test (white on black)" : name);
  palSel.appendChild(opt);
});

function updateCoverageLabel() {
  const v = document.getElementById("coverage").valueAsNumber;
  document.getElementById("coverageVal").textContent = v < 0 ? "auto" : v.toFixed(2);
}

function buildQuery() {
  const seed = document.getElementById("seed").value;
  const score = document.getElementById("score").value;
  const grid = document.getElementById("grid").value;
  const sym = document.getElementById("sym").value;
  const cov = document.getElementById("coverage").valueAsNumber;
  const sl = document.getElementById("strand_length").value;
  const pal = document.getElementById("palette").value;
  const ms = document.getElementById("min_score").valueAsNumber;
  const sb = document.getElementById("score_bands").value;
  const params = new URLSearchParams();
  params.set("seed", seed);
  params.set("score", score);
  params.set("grid", grid);
  if (sym !== "auto") params.set("sym", sym);
  if (cov >= 0) params.set("coverage", cov.toFixed(2));
  if (sl !== "auto") params.set("strand_length", sl);
  if (pal !== "auto") params.set("palette", pal);
  if (ms > 0) params.set("min_score", ms.toFixed(2));
  if (sb !== "") params.set("score_bands", sb);
  return params;
}

async function render() {
  const params = buildQuery();
  const url = "./render?" + params.toString();
  const preview = document.getElementById("preview");
  preview.innerHTML = '<span style="color: #aaa;">rendering…</span>';
  try {
    const resp = await fetch(url);
    if (!resp.ok) {
      preview.innerHTML = '<span style="color: #c33;">' + (await resp.text()) + '</span>';
      return;
    }
    const svg = await resp.text();
    preview.innerHTML = svg;
    const group = resp.headers.get("X-Kolavatar-Group") || "?";
    const palette = resp.headers.get("X-Kolavatar-Palette") || "?";
    const tier = resp.headers.get("X-Kolavatar-Tier") || "?";
    const cells = resp.headers.get("X-Kolavatar-Cells-Covered") || "?";
    const strands = resp.headers.get("X-Kolavatar-Strands") || "?";
    const seedHash = resp.headers.get("X-Kolavatar-Seed-Hash") || "?";
    document.getElementById("meta").innerHTML =
      '<span>group <code>' + group + '</code></span>' +
      '<span>palette <code>' + palette + '</code></span>' +
      '<span>tier <code>' + tier + '</code></span>' +
      '<span>cells <code>' + cells + '</code></span>' +
      '<span>strands <code>' + strands + '</code></span>' +
      '<span title="full seed hash">hash <code>' + seedHash.slice(0, 8) + '</code></span>';
    const shareURL = window.location.origin + window.location.pathname + "#" + params.toString();
    document.getElementById("share-url").value = shareURL;
  } catch (e) {
    preview.innerHTML = '<span style="color: #c33;">' + e.message + '</span>';
  }
}

function downloadSVG() {
  const svg = document.querySelector("#preview svg");
  if (!svg) return;
  const blob = new Blob([svg.outerHTML], {type: "image/svg+xml"});
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "kolavatar-" + document.getElementById("seed").value + ".svg";
  a.click();
  URL.revokeObjectURL(url);
}

async function randomizeSeed() {
  const r = await fetch("./random-seed");
  document.getElementById("seed").value = await r.text();
  render();
}

if (window.location.hash) {
  const params = new URLSearchParams(window.location.hash.slice(1));
  const setIf = (id, key) => { if (params.has(key)) document.getElementById(id).value = params.get(key); };
  setIf("seed", "seed");
  setIf("score", "score");
  setIf("grid", "grid");
  setIf("sym", "sym");
  setIf("strand_length", "strand_length");
  setIf("palette", "palette");
  setIf("score_bands", "score_bands");
  if (params.has("coverage")) document.getElementById("coverage").value = params.get("coverage");
  if (params.has("min_score")) document.getElementById("min_score").value = params.get("min_score");
  document.getElementById("scoreVal").textContent = document.getElementById("score").value;
  document.getElementById("minScoreVal").textContent = parseFloat(document.getElementById("min_score").value).toFixed(2);
  updateCoverageLabel();
}

["seed", "score", "grid", "sym", "coverage", "strand_length", "palette", "min_score", "score_bands"].forEach(id => {
  document.getElementById(id).addEventListener("change", render);
});
document.getElementById("score").addEventListener("input", debounce(render, 120));
document.getElementById("coverage").addEventListener("input", debounce(render, 120));
document.getElementById("min_score").addEventListener("input", debounce(render, 120));

function debounce(fn, ms) {
  let t;
  return (...args) => {
    clearTimeout(t);
    t = setTimeout(() => fn(...args), ms);
  };
}

render();
</script>
</body>
</html>
`
