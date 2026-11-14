<script setup lang="ts">
import { computed, ref } from "vue";
import HistogramChart from "./HistogramChart.vue";
import VariableStats from "./VariableStats.vue";
import type {
  DiscreteRequest,
  DiscreteResult,
} from "../services/randomVariablesService";
import { simulateDiscrete } from "../services/randomVariablesService";
import type { HistogramBar } from "../types/randomVariables";
import { formatCount, formatNumber } from "../utils/format";

interface Row {
  value: number;
  probability: number;
}

const rows = ref<Row[]>([
  { value: 1, probability: 0.1 },
  { value: 2, probability: 0.2 },
  { value: 3, probability: 0.3 },
  { value: 4, probability: 0.25 },
  { value: 5, probability: 0.15 },
]);
const sampleSize = ref(10000);
const alpha = ref(0.05);

const result = ref<DiscreteResult | null>(null);
const loading = ref(false);
const error = ref("");

const probabilitySum = computed(() =>
  rows.value.reduce((sum, row) => sum + row.probability, 0),
);
const sumValid = computed(
  () => Math.abs(probabilitySum.value - 1) <= 1e-6,
);

function addRow() {
  const maxValue = rows.value.length
    ? Math.max(...rows.value.map((row) => row.value))
    : 0;
  rows.value.push({ value: maxValue + 1, probability: 0.1 });
}

function removeRow(index: number) {
  if (rows.value.length > 2) {
    rows.value.splice(index, 1);
  }
}

function normalize() {
  const sum = probabilitySum.value;
  if (sum <= 0) return;

  const scale = 1e6;
  const scaled = rows.value.map((row) =>
    Math.round((row.probability / sum) * scale),
  );
  const largest = scaled.indexOf(Math.max(...scaled));
  scaled[largest] += scale - scaled.reduce((total, value) => total + value, 0);
  rows.value.forEach((row, index) => {
    row.probability = scaled[index] / scale;
  });
}

function validate(): string {
  if (rows.value.length < 2) {
    return "Нужно не меньше двух значений y_i.";
  }
  if (rows.value.some((row) => !Number.isFinite(row.value))) {
    return "Все значения y_i должны быть числами.";
  }
  if (rows.value.some((row) => !(row.probability > 0))) {
    return "Все вероятности p_i должны быть положительными.";
  }
  if (!sumValid.value) {
    return "Сумма вероятностей должна быть равна 1.";
  }
  const distinct = new Set(rows.value.map((row) => row.value));
  if (distinct.size !== rows.value.length) {
    return "Значения y_i не должны повторяться.";
  }
  if (sampleSize.value < 10 || sampleSize.value > 1000000) {
    return "N должно быть от 10 до 1 000 000.";
  }
  if (alpha.value < 0.000001 || alpha.value >= 0.5) {
    return "Уровень значимости α должен быть в интервале [0,000001; 0,5).";
  }
  return "";
}

const histogramBars = computed<HistogramBar[]>(() =>
  (result.value?.categories ?? []).map((category) => ({
    label: `y = ${formatNumber(category.value)}`,
    title: `y = ${formatNumber(category.value)}: наблюдение ${formatCount(category.count)}, теория ${formatCount(category.expectedCount)}`,
    observed: category.count,
    expected: category.expectedCount,
  })),
);

const preview = computed(() =>
  (result.value?.preview ?? [])
    .slice(0, 24)
    .map((value) => formatNumber(value))
    .join(", "),
);

function deltaLabel(index: number): string {
  const categories = result.value?.categories ?? [];
  let left = 0;
  for (let i = 0; i < index; i++) {
    left += categories[i].probability;
  }
  const right = left + categories[index].probability;
  return `(${formatNumber(left, 4)}; ${formatNumber(right, 4)}]`;
}

async function run() {
  error.value = "";
  const validation = validate();
  if (validation) {
    error.value = validation;
    return;
  }

  const request: DiscreteRequest = {
    values: rows.value.map((row) => row.value),
    probabilities: rows.value.map((row) => row.probability),
    n: sampleSize.value,
    alpha: alpha.value,
  };

  loading.value = true;
  try {
    result.value = await simulateDiscrete(request);
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
      <h2>Дискретная случайная величина</h2>
      <p>
        Интервал (0, 1) делится на участки Δᵢ длиной pᵢ. Равномерное число X
        попадает в один из участков, и величина принимает значение yᵢ.
      </p>
    </header>

    <div class="category-editor discrete-editor">
      <div class="table-wrap">
        <table class="category-table">
          <thead>
            <tr>
              <th>i</th>
              <th>Значение yᵢ</th>
              <th>Вероятность pᵢ</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, index) in rows" :key="index">
              <td>{{ index + 1 }}</td>
              <td>
                <input v-model.number="row.value" type="number" step="0.1" />
              </td>
              <td>
                <input
                  v-model.number="row.probability"
                  type="number"
                  min="0.000001"
                  max="1"
                  step="0.05"
                />
              </td>
              <td>
                <button
                  class="btn-small"
                  type="button"
                  :disabled="rows.length <= 2"
                  @click="removeRow(index)"
                >
                  Удалить
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="category-actions">
        <button class="btn-small" type="button" @click="addRow">
          Добавить значение
        </button>
        <button class="btn-small" type="button" @click="normalize">
          Нормировать
        </button>
        <span :class="['probability-sum', sumValid ? 'ok' : 'bad']">
          Σ<var>p</var> = {{ probabilitySum.toFixed(6) }}
        </span>
      </div>
    </div>

    <div class="controls-grid">
      <label class="field">
        <span>N</span>
        <input v-model.number="sampleSize" type="number" min="10" max="1000000" />
      </label>
      <label class="field">
        <span><var>α</var></span>
        <input v-model.number="alpha" type="number" min="0.001" max="0.49" step="0.01" />
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
          <h3>Дискретная величина</h3>
          <span class="status-chip">N = {{ formatCount(result.sampleSize) }}</span>
        </div>
        <p class="sample-preview">
          <strong>Первые значения:</strong> {{ preview }}…
        </p>
      </article>

      <HistogramChart
        :bars="histogramBars"
        title="Гистограмма и ожидаемые частоты"
        :rotate-labels="false"
      />

      <VariableStats
        :sample-size="result.sampleSize"
        :stats="result.stats"
        :theoretical-mean="result.theoreticalMean"
        :theoretical-variance="result.theoreticalVariance"
        :mean-ci="result.meanCi"
        :variance-ci="result.varianceCi"
        :test="result.test"
      />

      <article class="result-card">
        <h3>Частоты значений</h3>
        <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>yᵢ</th>
                <th>Отрезок Δᵢ</th>
                <th>Наблюдений</th>
                <th>Частота</th>
                <th>Вероятность pᵢ</th>
                <th>Ожидаемо</th>
                <th>Отклонение</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(category, index) in result.categories ?? []"
                :key="index"
              >
                <td>{{ formatNumber(category.value) }}</td>
                <td>{{ deltaLabel(index) }}</td>
                <td>{{ formatCount(category.count) }}</td>
                <td>{{ formatNumber(category.frequency, 5) }}</td>
                <td>{{ formatNumber(category.probability, 6) }}</td>
                <td>{{ formatCount(category.expectedCount) }}</td>
                <td>
                  {{ formatNumber(category.count - category.expectedCount, 1) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </article>
    </template>
  </section>
</template>
