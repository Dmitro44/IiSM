<script setup lang="ts">
import { computed } from "vue";
import type { HistogramBar } from "../types/randomVariables";
import { formatCount } from "../utils/format";

const props = withDefaults(
  defineProps<{
    bars: HistogramBar[];
    title?: string;
    height?: number;
    rotateLabels?: boolean;
  }>(),
  {
    title: "Гистограмма",
    height: 220,
    rotateLabels: true,
  },
);

const maxValue = computed(() =>
  Math.max(1, ...props.bars.map((bar) => Math.max(bar.observed, bar.expected))),
);

const axisMax = computed(() => maxValue.value);

function heightPercent(value: number): string {
  if (value <= 0) return "0%";
  return `${Math.max((value / axisMax.value) * 100, 1)}%`;
}
</script>

<template>
  <div class="histogram-card">
    <div class="histogram-head">
      <h4>{{ title }}</h4>
      <div class="histogram-legend">
        <span><i class="legend-swatch observed"></i> наблюдение</span>
        <span><i class="legend-swatch expected"></i> теория</span>
      </div>
    </div>

    <div class="histogram-scroll">
      <div class="histogram-plot" :style="{ '--chart-height': `${height}px` }">
        <div class="histogram-axis" aria-hidden="true">
          <span>{{ formatCount(axisMax) }}</span>
          <span>0</span>
        </div>

        <div class="histogram-bars">
          <div
            v-for="(bar, index) in bars"
            :key="index"
            class="histogram-column"
            :class="{ rotated: rotateLabels }"
            :title="bar.title ?? `${bar.label}: наблюдение ${formatCount(bar.observed)}, теория ${formatCount(bar.expected)}`"
          >
            <div class="histogram-bar-group">
              <div
                class="histogram-bar observed"
                :style="{ height: heightPercent(bar.observed) }"
              ></div>
              <div
                class="histogram-bar expected"
                :style="{ height: heightPercent(bar.expected) }"
              ></div>
            </div>
            <span class="histogram-label">{{ bar.label }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
