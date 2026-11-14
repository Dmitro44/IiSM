<script setup lang="ts">
import { computed, ref } from "vue";
import HistogramChart from "./HistogramChart.vue";
import VariableStats from "./VariableStats.vue";
import type {
  ContinuousRequest,
  ContinuousResult,
} from "../services/randomVariablesService";
import { simulateContinuous } from "../services/randomVariablesService";
import type { HistogramBar } from "../types/randomVariables";
import { formatCount, formatNumber } from "../utils/format";

type Law = "uniform" | "exponential";

const law = ref<Law>("exponential");
const a = ref(0);
const b = ref(1);
const lambda = ref(1);
const sampleSize = ref(10000);
const bins = ref(10);
const alpha = ref(0.05);

const result = ref<ContinuousResult | null>(null);
const loading = ref(false);
const error = ref("");

const formulas: Record<Law, string> = {
  uniform: "y = a + (b − a)·x",
  exponential: "y = −ln(1 − x) / λ",
};

const formula = computed(() => formulas[law.value]);

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
  if (law.value === "uniform" && a.value >= b.value) {
    return "Для равномерного закона нужно a < b.";
  }
  if (law.value === "exponential" && lambda.value <= 0) {
    return "Параметр λ должен быть положительным.";
  }
  return "";
}

function intervalLabel(bin: {
  lower: number;
  upper: number;
  lowerUnbounded: boolean;
  upperUnbounded: boolean;
}): string {
  const left = bin.lowerUnbounded ? "−∞" : formatNumber(bin.lower, 2);
  const right = bin.upperUnbounded ? "+∞" : formatNumber(bin.upper, 2);
  return `${left}…${right}`;
}

const histogramBars = computed<HistogramBar[]>(() =>
  (result.value?.histogram ?? []).map((bin) => ({
    label: intervalLabel(bin),
    title: `[${intervalLabel(bin)}]: наблюдение ${formatCount(bin.count)}, теория ${formatCount(bin.expectedCount)}`,
    observed: bin.count,
    expected: bin.expectedCount,
  })),
);

const preview = computed(() =>
  (result.value?.preview ?? [])
    .slice(0, 16)
    .map((value) => formatNumber(value, 4))
    .join(", "),
);

async function run() {
  error.value = "";
  const validation = validate();
  if (validation) {
    error.value = validation;
    return;
  }

  const request: ContinuousRequest = {
    law: law.value,
    a: a.value,
    b: b.value,
    lambda: lambda.value,
    n: sampleSize.value,
    bins: bins.value,
    alpha: alpha.value,
  };

  loading.value = true;
  try {
    result.value = await simulateContinuous(request);
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
      <h2>Непрерывная случайная величина</h2>
      <p>
        Метод обратных функций: из равномерного числа X на [0, 1) получается
        значение Y = G⁻¹(X) с заданной функцией распределения.
      </p>
    </header>

    <div class="formula-strip">
      <code>Y = G⁻¹(X)</code>
      <span>{{ formula }}</span>
    </div>

    <div class="controls-grid">
      <label class="field">
        <span>Закон распределения</span>
        <select v-model="law">
          <option value="uniform">Равномерный</option>
          <option value="exponential">Экспоненциальный</option>
        </select>
      </label>

      <template v-if="law === 'uniform'">
        <label class="field">
          <span><var>a</var></span>
          <input v-model.number="a" type="number" step="0.1" />
        </label>
        <label class="field">
          <span><var>b</var></span>
          <input v-model.number="b" type="number" step="0.1" />
        </label>
      </template>

      <label v-if="law === 'exponential'" class="field">
        <span><var>λ</var></span>
        <input v-model.number="lambda" type="number" min="0.0001" step="0.1" />
      </label>

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
          min="0.001"
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
          <h3>{{ result.name }}</h3>
          <span class="status-chip"
            >N = {{ formatCount(result.sampleSize) }}</span
          >
        </div>
        <p class="result-note">
          Обратное преобразование: <code>{{ result.formula }}</code>
        </p>
        <p class="sample-preview">
          <strong>Первые значения:</strong> {{ preview }}…
        </p>
      </article>

      <HistogramChart
        :bars="histogramBars"
        title="Гистограмма и ожидаемые частоты"
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
        <h3>Частоты по интервалам</h3>
        <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Интервал</th>
                <th>Наблюдений</th>
                <th>Частота</th>
                <th>Ожидаемо</th>
                <th>Отклонение</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(bin, index) in result.histogram ?? []" :key="index">
                <td>{{ intervalLabel(bin) }}</td>
                <td>{{ formatCount(bin.count) }}</td>
                <td>{{ formatNumber(bin.frequency, 5) }}</td>
                <td>{{ formatCount(bin.expectedCount) }}</td>
                <td>
                  {{ formatNumber(bin.count - bin.expectedCount, 1) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </article>
    </template>
  </section>
</template>
