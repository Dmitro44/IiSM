<script setup lang="ts">
import { computed, ref } from "vue";
import BivariateStats from "./BivariateStats.vue";
import HistogramChart from "./HistogramChart.vue";
import type { ConditionalRow } from "../../bindings/github.com/Dmitro44/IiSM/internal/bivariate/models";
import type {
  DiscreteRequest,
  DiscreteResult,
} from "../services/bivariateService";
import { simulateDiscrete } from "../services/bivariateService";
import type { HistogramBar } from "../types/randomVariables";
import { formatCount, formatNumber } from "../utils/format";

const sampleSize = ref(10000);
const alpha = ref(0.05);

const result = ref<DiscreteResult | null>(null);
const loading = ref(false);
const error = ref("");
const selectedXIndex = ref(0);
const selectedYIndex = ref(0);

function validate(): string {
  if (sampleSize.value < 10 || sampleSize.value > 1000000) {
    return "N должно быть от 10 до 1 000 000.";
  }
  if (alpha.value < 0.000001 || alpha.value >= 0.5) {
    return "Уровень значимости α должен быть в интервале [0,000001; 0,5).";
  }
  return "";
}

function cellValue(
  rows: (number[] | null)[],
  i: number,
  j: number,
): number {
  return rows[i]?.[j] ?? 0;
}

function rowSum(rows: (number[] | null)[], i: number): number {
  return (rows[i] ?? []).reduce(
    (sum, value) => sum + (Number.isFinite(value) ? value : 0),
    0,
  );
}

function columnSum(rows: (number[] | null)[], j: number): number {
  let sum = 0;
  for (const row of rows) {
    const value = row?.[j];
    if (value !== undefined && Number.isFinite(value)) {
      sum += value;
    }
  }
  return sum;
}

const valuesX = computed(() => result.value?.valuesX ?? []);
const valuesY = computed(() => result.value?.valuesY ?? []);
const probabilities = computed(() => result.value?.probabilities ?? []);
const observed = computed(() => result.value?.observed ?? []);

const probabilityTotal = computed(() => {
  let total = 0;
  for (let i = 0; i < probabilities.value.length; i++) {
    total += rowSum(probabilities.value, i);
  }
  return total;
});

const observedTotal = computed(() => {
  let total = 0;
  for (let i = 0; i < observed.value.length; i++) {
    total += rowSum(observed.value, i);
  }
  return total;
});

const preview = computed(() =>
  (result.value?.preview ?? [])
    .slice(0, 12)
    .map(
      (point) =>
        `(${formatNumber(point.x, 4)}; ${formatNumber(point.y, 4)})`,
    )
    .join("; "),
);

const histogramXBars = computed<HistogramBar[]>(() =>
  (result.value?.histogramX ?? []).map((category) => ({
    label: `x = ${formatNumber(category.value)}`,
    title: `x = ${formatNumber(category.value)}: наблюдение ${formatCount(category.count)}, ожидаемо ${formatCount(category.expectedCount)}`,
    observed: category.count,
    expected: category.expectedCount,
  })),
);

const histogramYBars = computed<HistogramBar[]>(() =>
  (result.value?.histogramY ?? []).map((category) => ({
    label: `y = ${formatNumber(category.value)}`,
    title: `y = ${formatNumber(category.value)}: наблюдение ${formatCount(category.count)}, ожидаемо ${formatCount(category.expectedCount)}`,
    observed: category.count,
    expected: category.expectedCount,
  })),
);

const conditionalXGivenY = computed<ConditionalRow[]>(
  () => result.value?.conditionalXGivenY ?? [],
);

const conditionalYGivenX = computed<ConditionalRow[]>(
  () => result.value?.conditionalYGivenX ?? [],
);

const selectedXConditional = computed<ConditionalRow | null>(
  () => conditionalXGivenY.value[selectedXIndex.value] ?? null,
);

const selectedYConditional = computed<ConditionalRow | null>(
  () => conditionalYGivenX.value[selectedYIndex.value] ?? null,
);

interface ConditionalDisplayRow {
  value: number;
  theoretical: number;
  observed: number;
}

function displayRows(selected: ConditionalRow | null): ConditionalDisplayRow[] {
  if (!selected) return [];
  return (selected.values ?? []).map((value, index) => ({
    value,
    theoretical: selected.theoretical?.[index] ?? 0,
    observed: selected.observed?.[index] ?? 0,
  }));
}

const xGivenYRows = computed<ConditionalDisplayRow[]>(() =>
  displayRows(selectedXConditional.value),
);

const yGivenXRows = computed<ConditionalDisplayRow[]>(() =>
  displayRows(selectedYConditional.value),
);

async function run() {
  error.value = "";
  const validation = validate();
  if (validation) {
    error.value = validation;
    return;
  }

  const request: DiscreteRequest = {
    n: sampleSize.value,
    alpha: alpha.value,
  };

  loading.value = true;
  try {
    result.value = await simulateDiscrete(request);
    selectedXIndex.value = 0;
    selectedYIndex.value = 0;
  } catch (err) {
    result.value = null;
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <section class="panel">
    <header class="panel-header">
      <h2>Двумерная дискретная величина</h2>
      <p>
        Матрица распределения задана в Go: X ∈ {1,2,3}, Y ∈ {0,1}. Центр ячейки
        выбирается по накопленной вероятности, затем выдаются значения X и Y.
      </p>
    </header>

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
        <p class="sample-preview">
          <strong>Первые точки:</strong> {{ preview }}
        </p>
      </article>

      <div class="matrix-grid">
        <article class="result-card">
          <h3>Матрица вероятностей <var>p<sub>ij</sub></var></h3>
          <div class="table-wrap">
            <table>
              <thead>
                <tr>
                  <th><var>X</var> \ <var>Y</var></th>
                  <th v-for="y in valuesY" :key="y">
                    <var>y</var> = {{ formatNumber(y) }}
                  </th>
                  <th>Σ</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(x, i) in valuesX" :key="x">
                  <td><var>x</var> = {{ formatNumber(x) }}</td>
                  <td v-for="(y, j) in valuesY" :key="y">
                    {{ formatNumber(cellValue(probabilities, i, j), 3) }}
                  </td>
                  <td>{{ formatNumber(rowSum(probabilities, i), 3) }}</td>
                </tr>
                <tr>
                  <td>Σ</td>
                  <td v-for="(y, j) in valuesY" :key="y">
                    {{ formatNumber(columnSum(probabilities, j), 3) }}
                  </td>
                  <td>{{ formatNumber(probabilityTotal, 3) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <p class="result-note">
            Сумма всех <var>p<sub>ij</sub></var> равна 1.
          </p>
        </article>

        <article class="result-card">
          <h3>Наблюдённые частоты <var>n<sub>ij</sub></var></h3>
          <div class="table-wrap">
            <table>
              <thead>
                <tr>
                  <th><var>X</var> \ <var>Y</var></th>
                  <th v-for="y in valuesY" :key="y">
                    <var>y</var> = {{ formatNumber(y) }}
                  </th>
                  <th>Σ</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(x, i) in valuesX" :key="x">
                  <td><var>x</var> = {{ formatNumber(x) }}</td>
                  <td v-for="(y, j) in valuesY" :key="y">
                    {{ formatCount(cellValue(observed, i, j)) }}
                  </td>
                  <td>{{ formatCount(rowSum(observed, i)) }}</td>
                </tr>
                <tr>
                  <td>Σ</td>
                  <td v-for="(y, j) in valuesY" :key="y">
                    {{ formatCount(columnSum(observed, j)) }}
                  </td>
                  <td>{{ formatCount(observedTotal) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </article>
      </div>

      <div class="chart-pair">
        <HistogramChart
          :bars="histogramXBars"
          title="Компонента X: наблюдаемые и ожидаемые частоты"
          :rotate-labels="false"
        />
        <HistogramChart
          :bars="histogramYBars"
          title="Компонента Y: наблюдаемые и ожидаемые частоты"
          :rotate-labels="false"
        />
      </div>

      <div class="conditional-grid">
        <article class="result-card">
          <h3>X при условии Y = y<sub>j</sub></h3>
          <template v-if="selectedXConditional">
            <label class="field conditional-field">
              <span>Значение y</span>
              <select v-model.number="selectedXIndex">
                <option
                  v-for="(row, index) in conditionalXGivenY"
                  :key="index"
                  :value="index"
                >
                  y = {{ formatNumber(row.fixed) }}
                </option>
              </select>
            </label>
            <div v-if="xGivenYRows.length" class="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th><var>x<sub>i</sub></var></th>
                    <th>
                      <var>P</var>(<var>X</var> = <var>x<sub>i</sub></var> |
                      <var>Y</var> = <var>y<sub>j</sub></var>)
                    </th>
                    <th>Наблюдаемая частота</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, index) in xGivenYRows" :key="index">
                    <td>{{ formatNumber(row.value) }}</td>
                    <td>{{ formatNumber(row.theoretical, 4) }}</td>
                    <td>{{ formatNumber(row.observed, 4) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <p v-else class="result-note">
              Условное распределение не построено.
            </p>
          </template>
          <p v-else class="result-note">
            Условные распределения не построены.
          </p>
        </article>

        <article class="result-card">
          <h3>Y при условии X = x<sub>i</sub></h3>
          <template v-if="selectedYConditional">
            <label class="field conditional-field">
              <span>Значение x</span>
              <select v-model.number="selectedYIndex">
                <option
                  v-for="(row, index) in conditionalYGivenX"
                  :key="index"
                  :value="index"
                >
                  x = {{ formatNumber(row.fixed) }}
                </option>
              </select>
            </label>
            <div v-if="yGivenXRows.length" class="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th><var>y<sub>j</sub></var></th>
                    <th>
                      <var>P</var>(<var>Y</var> = <var>y<sub>j</sub></var> |
                      <var>X</var> = <var>x<sub>i</sub></var>)
                    </th>
                    <th>Наблюдаемая частота</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, index) in yGivenXRows" :key="index">
                    <td>{{ formatNumber(row.value) }}</td>
                    <td>{{ formatNumber(row.theoretical, 4) }}</td>
                    <td>{{ formatNumber(row.observed, 4) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <p v-else class="result-note">
              Условное распределение не построено.
            </p>
          </template>
          <p v-else class="result-note">
            Условные распределения не построены.
          </p>
        </article>
      </div>

      <BivariateStats
        :sample-size="result.sampleSize"
        :stats="result.stats"
        :theoretical="result.theoretical"
        :intervals="result.intervals"
        :tests="result.tests ?? []"
      />

      <p class="result-note">
        Компоненты зависимы: <var>P</var>(<var>X</var> = <var>x<sub>i</sub></var>,
        <var>Y</var> = <var>y<sub>j</sub></var>) ≠
        <var>P</var>(<var>X</var> = <var>x<sub>i</sub></var>)·<var>P</var>(<var>Y</var>
        = <var>y<sub>j</sub></var>), поэтому критерий независимости гипотезу
        отвергает. Теоретическая корреляция при этом не нулевая.
      </p>
    </template>
  </section>
</template>

<style scoped>
.chart-pair,
.matrix-grid,
.conditional-grid {
  display: grid;
  gap: var(--s-3);
  margin-top: var(--s-3);
}

.chart-pair,
.matrix-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.conditional-grid {
  grid-template-columns: repeat(auto-fit, minmax(340px, 1fr));
}

.chart-pair > *,
.matrix-grid > .result-card,
.conditional-grid > .result-card {
  margin-top: 0;
  min-width: 0;
}

.conditional-field {
  max-width: 16rem;
}

@media (max-width: 1000px) {
  .chart-pair,
  .matrix-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
