<script setup lang="ts">
import { computed } from "vue";
import type { DensityBin } from "../types/bivariateUi";
import { formatNumber } from "../utils/format";

const props = withDefaults(
  defineProps<{
    bins: DensityBin[];
    title?: string;
    xLabel?: string;
    yLabel?: string;
  }>(),
  {
    title: "Плотность распределения",
    xLabel: "x",
    yLabel: "f(x)",
  },
);

const W = 640;
const H = 340;
const PAD = { top: 28, right: 24, bottom: 66, left: 68 };
const plotW = W - PAD.left - PAD.right;
const plotH = H - PAD.top - PAD.bottom;
const baseY = PAD.top + plotH;

function safe(value: number): number {
  return Number.isFinite(value) ? value : 0;
}

const domain = computed(() => {
  const edges: number[] = [];
  for (const bin of props.bins) {
    edges.push(safe(bin.lower), safe(bin.upper));
  }
  if (edges.length === 0) {
    return { min: 0, max: 1 };
  }
  let min = Math.min(...edges);
  let max = Math.max(...edges);
  if (!(max > min)) {
    min -= 0.5;
    max += 0.5;
  }
  return { min, max };
});

const yMax = computed(() => {
  let max = 0;
  for (const bin of props.bins) {
    max = Math.max(max, safe(bin.density), safe(bin.theoreticalDensity));
  }
  return max > 0 ? max : 1;
});

function scaleX(value: number): number {
  const { min, max } = domain.value;
  const ratio = (safe(value) - min) / (max - min);
  return PAD.left + Math.min(Math.max(ratio, 0), 1) * plotW;
}

function scaleY(value: number): number {
  const ratio = Math.max(safe(value), 0) / yMax.value;
  return baseY - ratio * plotH;
}

const bars = computed(() =>
  props.bins.map((bin) => {
    const lower = safe(bin.lower);
    const upper = safe(bin.upper);
    const xa = scaleX(Math.min(lower, upper));
    const xb = scaleX(Math.max(lower, upper));
    const width = Math.max(xb - xa - 1.5, 0.8);
    const x = xa + (xb - xa - width) / 2;
    const density = Math.max(safe(bin.density), 0);
    const y = scaleY(density);
    return {
      x,
      y,
      width,
      height: Math.max(baseY - y, 0),
      title: `[${formatNumber(lower, 3)}; ${formatNumber(upper, 3)}]: n = ${formatNumber(
        bin.count,
        0,
      )}, плотность = ${formatNumber(density, 5)}`,
    };
  }),
);

const curveDots = computed(() =>
  props.bins.map((bin) => ({
    x: scaleX((safe(bin.lower) + safe(bin.upper)) / 2),
    y: scaleY(safe(bin.theoreticalDensity)),
  })),
);

const curvePoints = computed(() =>
  curveDots.value
    .map((point) => `${point.x.toFixed(2)},${point.y.toFixed(2)}`)
    .join(" "),
);

const yTicks = computed(() =>
  [0, 0.25, 0.5, 0.75, 1].map((fraction) => ({
    y: baseY - fraction * plotH,
    label: formatNumber(yMax.value * fraction, 3),
  })),
);

const xTicks = computed(() => {
  const { min, max } = domain.value;
  return [
    { x: scaleX(min), label: formatNumber(min, 3), anchor: "start" },
    { x: scaleX((min + max) / 2), label: formatNumber((min + max) / 2, 3), anchor: "middle" },
    { x: scaleX(max), label: formatNumber(max, 3), anchor: "end" },
  ];
});

const allZero = computed(
  () =>
    props.bins.length > 0 &&
    props.bins.every(
      (bin) => safe(bin.density) <= 0 && safe(bin.theoreticalDensity) <= 0,
    ),
);
</script>

<template>
  <figure class="density-chart">
    <figcaption class="chart-head">
      <h4>{{ title }}</h4>
      <div class="chart-legend">
        <span><i class="swatch swatch-bar"></i>плотность по выборке</span>
        <span><i class="swatch swatch-line"></i>теоретическая плотность</span>
      </div>
    </figcaption>

    <div class="chart-body">
      <svg :viewBox="`0 0 ${W} ${H}`" role="img" :aria-label="title">
        <g class="grid">
          <line
            v-for="tick in yTicks"
            :key="`grid-${tick.y}`"
            :x1="PAD.left"
            :x2="PAD.left + plotW"
            :y1="tick.y"
            :y2="tick.y"
          />
        </g>

        <g class="bars">
          <rect
            v-for="(bar, index) in bars"
            :key="`bar-${index}`"
            :x="bar.x"
            :y="bar.y"
            :width="bar.width"
            :height="bar.height"
            rx="1"
          >
            <title>{{ bar.title }}</title>
          </rect>
        </g>

        <g class="curve">
          <polyline v-if="curveDots.length > 1" :points="curvePoints" />
          <circle
            v-for="(dot, index) in curveDots"
            :key="`dot-${index}`"
            :cx="dot.x"
            :cy="dot.y"
            r="2.2"
          />
        </g>

        <g class="axes">
          <line :x1="PAD.left" :x2="PAD.left" :y1="PAD.top" :y2="baseY" />
          <line :x1="PAD.left" :x2="PAD.left + plotW" :y1="baseY" :y2="baseY" />
        </g>

        <g class="tick-labels">
          <text
            v-for="tick in yTicks"
            :key="`y-${tick.y}`"
            :x="PAD.left - 8"
            :y="tick.y + 3.5"
            text-anchor="end"
          >
            {{ tick.label }}
          </text>
          <text
            v-for="tick in xTicks"
            :key="`x-${tick.anchor}`"
            :x="tick.x"
            :y="baseY + 18"
            :text-anchor="tick.anchor"
          >
            {{ tick.label }}
          </text>
        </g>

        <text :x="PAD.left + plotW / 2" :y="H - 12" class="axis-label" text-anchor="middle">
          {{ xLabel }}
        </text>
        <text
          :x="16"
          :y="PAD.top + plotH / 2"
          class="axis-label"
          text-anchor="middle"
          :transform="`rotate(-90 16 ${PAD.top + plotH / 2})`"
        >
          {{ yLabel }}
        </text>
      </svg>
    </div>

    <p v-if="bins.length === 0" class="chart-note">Нет данных для построения гистограммы.</p>
    <p v-else-if="allZero" class="chart-note">
      Все плотности равны нулю — масштаб по оси f(x) показан условно.
    </p>
  </figure>
</template>

<style scoped>
.density-chart {
  margin: 0;
  padding: var(--s-2);
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius);
}

.chart-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 0.5rem 1rem;
}

.chart-head h4 {
  margin: 0;
  color: var(--text);
  font-size: 1rem;
}

.chart-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 0.9rem;
  color: var(--muted);
  font-size: 0.78rem;
}

.chart-legend span {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}

.swatch {
  display: inline-block;
  flex: 0 0 auto;
}

.swatch-bar {
  width: 0.7rem;
  height: 0.7rem;
  border-radius: 2px;
  background: var(--accent);
}

.swatch-line {
  width: 1.2rem;
  height: 2px;
  border-radius: 1px;
  background: var(--text);
}

.chart-body {
  margin-top: var(--s-1);
}

svg {
  display: block;
  width: 100%;
  height: auto;
  overflow: visible;
}

.grid line {
  stroke: rgba(255, 255, 255, 0.08);
  stroke-width: 1;
}

.bars rect {
  fill: var(--accent);
  fill-opacity: 0.55;
  stroke: var(--accent);
  stroke-opacity: 0.8;
  stroke-width: 1;
}

.curve polyline {
  fill: none;
  stroke: var(--text);
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.curve circle {
  fill: var(--text);
}

.axes line {
  stroke: rgba(255, 255, 255, 0.28);
  stroke-width: 1;
}

.tick-labels text {
  fill: var(--muted);
  font-size: 12px;
}

.axis-label {
  fill: var(--muted);
  font-size: 13px;
  letter-spacing: 0.04em;
}

.chart-note {
  margin: var(--s-1) 0 0;
  color: var(--muted);
  font-size: 0.78rem;
  line-height: 1.4;
}
</style>
