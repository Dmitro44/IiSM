<script setup lang="ts">
import { computed } from "vue";
import type { CurvePoint } from "../types/bivariateUi";
import { formatNumber } from "../utils/format";

const props = withDefaults(
  defineProps<{
    points: CurvePoint[];
    fixedX: number;
    marginalDensity: number;
    title?: string;
  }>(),
  {
    title: "Условная плотность",
  },
);

const W = 600;
const H = 300;
const PAD = { top: 26, right: 24, bottom: 62, left: 68 };
const plotW = W - PAD.left - PAD.right;
const plotH = H - PAD.top - PAD.bottom;
const baseY = PAD.top + plotH;

function safe(value: number): number {
  return Number.isFinite(value) ? value : 0;
}

const ordered = computed(() =>
  props.points
    .filter((point) => Number.isFinite(point.x) && Number.isFinite(point.y))
    .map((point) => ({ x: point.x, y: Math.max(point.y, 0) }))
    .sort((a, b) => a.x - b.x),
);

const domain = computed(() => {
  if (ordered.value.length === 0) {
    return { min: 0, max: 1 };
  }
  let min = Math.min(...ordered.value.map((point) => point.x));
  let max = Math.max(...ordered.value.map((point) => point.x));
  if (!(max > min)) {
    min -= 0.5;
    max += 0.5;
  }
  return { min, max };
});

const yMax = computed(() => {
  let max = 0;
  for (const point of ordered.value) {
    max = Math.max(max, point.y);
  }
  return max > 0 ? max : 1;
});

function scaleX(value: number): number {
  const { min, max } = domain.value;
  const ratio = (value - min) / (max - min);
  return PAD.left + Math.min(Math.max(ratio, 0), 1) * plotW;
}

function scaleY(value: number): number {
  return baseY - (Math.max(value, 0) / yMax.value) * plotH;
}

const curve = computed(() =>
  ordered.value.map((point) => ({
    x: scaleX(point.x),
    y: scaleY(point.y),
  })),
);

const linePoints = computed(() =>
  curve.value.map((point) => `${point.x.toFixed(2)},${point.y.toFixed(2)}`).join(" "),
);

const areaPath = computed(() => {
  if (curve.value.length === 0) return "";
  const segments = curve.value.map(
    (point, index) => `${index === 0 ? "M" : "L"} ${point.x.toFixed(2)} ${point.y.toFixed(2)}`,
  );
  const first = curve.value[0];
  const last = curve.value[curve.value.length - 1];
  return `${segments.join(" ")} L ${last.x.toFixed(2)} ${baseY.toFixed(2)} L ${first.x.toFixed(
    2,
  )} ${baseY.toFixed(2)} Z`;
});

const yTicks = computed(() =>
  [0, 0.25, 0.5, 0.75, 1].map((fraction) => ({
    y: baseY - fraction * plotH,
    label: formatNumber(yMax.value * fraction, 3),
  })),
);

const xTicks = computed(() => {
  const { min, max } = domain.value;
  return Array.from({ length: 5 }, (_, index) => {
    const value = min + ((max - min) * index) / 4;
    const anchor = index === 0 ? "start" : index === 4 ? "end" : "middle";
    return { x: scaleX(value), label: formatNumber(value, 3), anchor };
  });
});

const allZero = computed(
  () => ordered.value.length > 0 && ordered.value.every((point) => point.y <= 0),
);

const fixedXLabel = computed(() => formatNumber(safe(props.fixedX), 3));
const marginalLabel = computed(() => formatNumber(safe(props.marginalDensity)));
</script>

<template>
  <figure class="conditional-chart">
    <figcaption class="chart-head">
      <h4>{{ title }}</h4>
      <span class="chart-tag">f(y | x = {{ fixedXLabel }})</span>
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

        <path v-if="areaPath" :d="areaPath" class="curve-area" />
        <polyline v-if="curve.length > 1" :points="linePoints" class="curve-line" />

        <line
          :x1="PAD.left"
          :x2="PAD.left + plotW"
          :y1="baseY"
          :y2="baseY"
          class="zero-line"
        />

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
            :key="`x-${tick.anchor}-${tick.x}`"
            :x="tick.x"
            :y="baseY + 18"
            :text-anchor="tick.anchor"
          >
            {{ tick.label }}
          </text>
        </g>

        <text :x="PAD.left + plotW / 2" :y="H - 12" class="axis-label" text-anchor="middle">
          y
        </text>
        <text
          :x="16"
          :y="PAD.top + plotH / 2"
          class="axis-label"
          text-anchor="middle"
          :transform="`rotate(-90 16 ${PAD.top + plotH / 2})`"
        >
          f(y | x)
        </text>
      </svg>
    </div>

    <div class="chart-foot">
      <p v-if="points.length === 0" class="chart-note">Нет данных для построения кривой.</p>
      <p v-else-if="allZero" class="chart-note">
        Условная плотность на всём диапазоне равна нулю.
      </p>
      <p class="chart-caption">
        <var>f<sub>X</sub></var>({{ fixedXLabel }}) = {{ marginalLabel }}
      </p>
    </div>
  </figure>
</template>

<style scoped>
.conditional-chart {
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

.chart-tag {
  padding: 0.16rem 0.55rem;
  border: 1px solid rgba(181, 189, 104, 0.28);
  border-radius: 999px;
  background: rgba(181, 189, 104, 0.1);
  color: var(--accent);
  font-size: 0.74rem;
  font-weight: 600;
  white-space: nowrap;
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

.curve-area {
  fill: var(--accent);
  fill-opacity: 0.16;
}

.curve-line {
  fill: none;
  stroke: var(--accent);
  stroke-width: 2.2;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.zero-line {
  stroke: var(--muted);
  stroke-opacity: 0.55;
  stroke-width: 1;
  stroke-dasharray: 4 4;
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

.chart-foot {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 0.35rem 1rem;
  margin-top: var(--s-1);
}

.chart-note {
  margin: 0;
  color: var(--muted);
  font-size: 0.78rem;
  line-height: 1.4;
}

.chart-caption {
  margin: 0 0 0 auto;
  color: var(--text);
  font-size: 0.9rem;
  font-weight: 600;
  letter-spacing: 0.01em;
}
</style>
