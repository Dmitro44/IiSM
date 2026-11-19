<script setup lang="ts">
import { computed, ref } from "vue";
import BivariateStats from "./BivariateStats.vue";
import ConditionalCurveChart from "./ConditionalCurveChart.vue";
import DensityHistogram from "./DensityHistogram.vue";
import JointHistogram3D from "./JointHistogram3D.vue";
import type { ConditionalCurve } from "../../bindings/github.com/Dmitro44/IiSM/internal/bivariate/models";
import type {
  ContinuousRequest,
  ContinuousResult,
} from "../services/bivariateService";
import { simulateContinuous } from "../services/bivariateService";
import type { CurvePoint } from "../types/bivariateUi";
import { formatCount, formatNumber } from "../utils/format";

const sampleSize = ref(10000);
const bins = ref(10);
const alpha = ref(0.05);

const result = ref<ContinuousResult | null>(null);
const loading = ref(false);
const error = ref("");
const selectedConditionalIndex = ref(0);

function validate(): string {
  if (sampleSize.value < 10 || sampleSize.value > 1000000) {
    return "N должно быть от 10 до 1 000 000.";
  }
  if (bins.value < 2 || bins.value > 50) {
    return "Число интервалов должно быть от 2 до 50.";
  }
  if (alpha.value < 0.000001 || alpha.value >= 0.5) {
    return "Уровень значимости α должен быть в интервале [0,000001; 0,5).";
  }
  return "";
}

function defaultConditionalIndex(curves: ConditionalCurve[]): number {
  const index = curves.findIndex((curve) => Math.abs(curve.fixedX) < 1e-9);
  return index >= 0 ? index : 0;
}

const preview = computed(() =>
  (result.value?.preview ?? [])
    .slice(0, 12)
    .map(
      (point) =>
        `(${formatNumber(point.x, 4)}; ${formatNumber(point.y, 4)})`,
    )
    .join("; "),
);

const conditionalCurves = computed<ConditionalCurve[]>(
  () => result.value?.conditional ?? [],
);

const selectedCurve = computed<ConditionalCurve | null>(
  () => conditionalCurves.value[selectedConditionalIndex.value] ?? null,
);

const correlationConclusion = computed(() => {
  const test = result.value?.tests?.find((item) => item.name === "Корреляция");
  if (!test?.valid) {
    return "Критерий корреляции неприменим, вывод о ρ сделать нельзя.";
  }
  return test.accepted
    ? "Критерий корреляции гипотезу ρ = 0 не отвергает."
    : "Критерий корреляции гипотезу ρ = 0 отвергает — при верной гипотезе это ошибка первого рода.";
});

const conditionalPoints = computed<CurvePoint[]>(() =>
  (selectedCurve.value?.points ?? []).map((point) => ({
    x: point.value,
    y: point.density,
  })),
);

async function run() {
  error.value = "";
  const validation = validate();
  if (validation) {
    error.value = validation;
    return;
  }

  const request: ContinuousRequest = {
    n: sampleSize.value,
    bins: bins.value,
    alpha: alpha.value,
  };

  loading.value = true;
  try {
    result.value = await simulateContinuous(request);
    selectedConditionalIndex.value = defaultConditionalIndex(
      result.value.conditional ?? [],
    );
  } catch (err) {
    result.value = null;
    selectedConditionalIndex.value = 0;
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <section class="panel">
    <header class="panel-header">
      <h2>Двумерная непрерывная величина</h2>
      <p>
        Вариант 3: <var>f(x,y)</var> = (9 − √(x²+y²))/(63π) при x²+y² ≤ 9.
        Точка равномерно выбирается в круге радиуса 3 и принимается с
        вероятностью (9−r)/9 — так плотность пропорциональна 9−r.
      </p>
    </header>

    <div class="formula-strip">
      <code>f(x, y) = (9 − √(x² + y²)) / (63π)</code>
      <span>при x² + y² ≤ 9</span>
    </div>

    <div class="controls-grid">
      <label class="field">
        <span>N</span>
        <input
          v-model.number="sampleSize"
          type="number"
          min="10"
          max="1000000"
        />
      </label>
      <label class="field">
        <span>Интервалов</span>
        <input v-model.number="bins" type="number" min="2" max="50" />
      </label>
      <label class="field">
        <span><var>α</var></span>
        <input
          v-model.number="alpha"
          type="number"
          min="0.000001"
          max="0.49"
          step="0.01"
        />
      </label>
    </div>

    <div class="actions">
      <button class="btn" :disabled="loading" @click="run">
        {{ loading ? "Расчёт..." : "Смоделировать" }}
      </button>
    </div>

    <p v-if="error" class="error">{{ error }}</p>

    <template v-if="result">
      <article class="result-card">
        <div class="result-head">
          <h3>Результат моделирования</h3>
          <span class="status-chip">
            N = {{ formatCount(result.sampleSize) }}
          </span>
        </div>
        <p class="result-note">
          Плотность: <code>{{ result.formula }}</code>
        </p>
        <p class="result-note">Метод: {{ result.method }}</p>
        <p class="sample-preview">
          <strong>Первые точки:</strong> {{ preview }}
        </p>
      </article>

      <div class="chart-pair">
        <DensityHistogram
          :bins="result.histogramX ?? []"
          title="Компонента X: гистограмма и f_X(x)"
          x-label="x"
          y-label="f_X(x)"
        />
        <DensityHistogram
          :bins="result.histogramY ?? []"
          title="Компонента Y: гистограмма и f_Y(y)"
          x-label="y"
          y-label="f_Y(y)"
        />
      </div>

      <article class="result-card">
        <h3>Условные плотности</h3>
        <p class="result-note">
          <var>f(y | x)</var> = <var>f(x,y)</var> / <var>f<sub>X</sub>(x)</var>.
          Маргинальная плотность получена аналитически интегрированием
          совместной по <var>y</var>:
        </p>
        <p class="result-note">
          <code>f_X(x) = 1/(63π) · [15a − (x²/2)·ln((3+a)/(3−a))]</code>,
          <var>a</var> = √(9 − <var>x</var>²).
        </p>

        <template v-if="selectedCurve">
          <label class="field conditional-field">
            <span>Значение x</span>
            <select v-model.number="selectedConditionalIndex">
              <option
                v-for="(curve, index) in conditionalCurves"
                :key="index"
                :value="index"
              >
                x = {{ formatNumber(curve.fixedX, 0) }}
              </option>
            </select>
          </label>

          <ConditionalCurveChart
            class="conditional-chart-block"
            :points="conditionalPoints"
            :fixed-x="selectedCurve.fixedX"
            :marginal-density="selectedCurve.marginalDensity"
          />
        </template>
        <p v-else class="result-note">Условные плотности не построены.</p>
      </article>

      <JointHistogram3D
        class="spaced-chart"
        :cells="result.joint ?? []"
        :grid-size="result.jointGridSize"
        title="Совместная гистограмма и теоретическая плотность (3D)"
      />

      <BivariateStats
        :sample-size="result.sampleSize"
        :stats="result.stats"
        :theoretical="result.theoretical"
        :intervals="result.intervals"
        :tests="result.tests ?? []"
      />

      <p class="result-note">
        Компоненты зависимы: носитель — круг, а не прямоугольник, поэтому
        совместная плотность не равна произведению маргинальных. При этом
        <var>r</var> = 0: зависимость нелинейная (радиальная).
        {{ correlationConclusion }}
      </p>
    </template>
  </section>
</template>

<style scoped>
.chart-pair {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--s-3);
  margin-top: var(--s-3);
}

@media (max-width: 1000px) {
  .chart-pair {
    grid-template-columns: minmax(0, 1fr);
  }
}

.conditional-field {
  max-width: 16rem;
  margin-top: var(--s-2);
}

.conditional-chart-block,
.spaced-chart {
  margin-top: var(--s-2);
}
</style>
