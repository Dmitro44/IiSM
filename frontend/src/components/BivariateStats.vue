<script setup lang="ts">
import { computed } from "vue";
import type {
  Interval,
  Intervals,
  SampleStats,
  Test,
  TheoreticalStats,
} from "../../bindings/github.com/Dmitro44/IiSM/internal/bivariate/models";
import { formatNumber, formatPValue } from "../utils/format";

const props = defineProps<{
  sampleSize: number;
  stats: SampleStats;
  theoretical: TheoreticalStats;
  intervals: Intervals;
  tests: Test[];
}>();

interface EstimateRow {
  label: string;
  sample: number;
  theory: number;
}

const estimateRows = computed<EstimateRow[]>(() => [
  { label: "M(X)", sample: props.stats.meanX, theory: props.theoretical.meanX },
  { label: "M(Y)", sample: props.stats.meanY, theory: props.theoretical.meanY },
  {
    label: "D(X)",
    sample: props.stats.varianceX,
    theory: props.theoretical.varianceX,
  },
  {
    label: "D(Y)",
    sample: props.stats.varianceY,
    theory: props.theoretical.varianceY,
  },
  {
    label: "cov(X,Y)",
    sample: props.stats.covariance,
    theory: props.theoretical.covariance,
  },
  {
    label: "r(X,Y)",
    sample: props.stats.correlation,
    theory: props.theoretical.correlation,
  },
]);

interface IntervalRow {
  label: string;
  interval: Interval;
}

const intervalRows = computed<IntervalRow[]>(() => [
  { label: "M(X)", interval: props.intervals.meanX },
  { label: "M(Y)", interval: props.intervals.meanY },
  { label: "D(X)", interval: props.intervals.varianceX },
  { label: "D(Y)", interval: props.intervals.varianceY },
  { label: "r(X,Y)", interval: props.intervals.correlation },
]);

function hasInterval(interval: Interval): boolean {
  return !(interval.lower === 0 && interval.upper === 0);
}
</script>

<template>
  <div class="bivariate-stats">
    <article class="result-card">
      <h3>Точечные оценки</h3>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Параметр</th>
              <th>Выборка</th>
              <th>Теория</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in estimateRows" :key="row.label">
              <td><var>{{ row.label }}</var></td>
              <td>{{ formatNumber(row.sample, 4) }}</td>
              <td>{{ formatNumber(row.theory, 4) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </article>

    <article class="result-card">
      <h3>Интервальные оценки</h3>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Параметр</th>
              <th>Нижняя граница</th>
              <th>Верхняя граница</th>
              <th>Доверительная вероятность</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in intervalRows" :key="row.label">
              <td><var>{{ row.label }}</var></td>
              <template v-if="hasInterval(row.interval)">
                <td>{{ formatNumber(row.interval.lower, 4) }}</td>
                <td>{{ formatNumber(row.interval.upper, 4) }}</td>
                <td>{{ formatNumber(row.interval.confidence, 3) }}</td>
              </template>
              <template v-else>
                <td title="интервал не построен">—</td>
                <td title="интервал не построен">—</td>
                <td>—</td>
              </template>
            </tr>
          </tbody>
        </table>
      </div>
    </article>

    <article class="result-card">
      <h3>Проверка гипотез</h3>

      <div v-for="test in tests ?? []" :key="test.name" class="test-card">
        <div class="test-head">
          <strong>{{ test.name }}</strong>
          <span
            :class="[
              'test-verdict',
              test.valid ? (test.accepted ? 'accepted' : 'rejected') : 'rejected',
            ]"
          >
            {{
              test.valid
                ? test.accepted
                  ? "Гипотеза не отвергается"
                  : "Гипотеза отвергается"
                : "Критерий неприменим"
            }}
          </span>
        </div>
        <div class="stats-grid test-stats">
          <div>
            <strong>Статистика</strong>
            <span>{{ formatNumber(test.statistic) }}</span>
          </div>
          <div>
            <strong>Степени свободы</strong>
            <span
              v-if="test.degreesOfFreedom === 0"
              title="для z-критерия число степеней свободы не используется"
            >
              —
            </span>
            <span v-else>{{ test.degreesOfFreedom }}</span>
          </div>
          <div>
            <strong><var>p</var>-значение</strong>
            <span>{{ formatPValue(test.pValue) }}</span>
          </div>
          <div>
            <strong>Уровень <var>α</var></strong>
            <span>{{ formatNumber(test.alpha, 3) }}</span>
          </div>
        </div>
        <p v-if="test.note" class="result-note">{{ test.note }}</p>
      </div>

      <p class="result-note">
        Среднее проверяется t-критерием Стьюдента, дисперсия — критерием
        <var>χ²</var>, корреляция — преобразованием Фишера. Критерий согласия —
        <var>χ²</var> Пирсона; для независимости строится таблица сопряжённости.
      </p>
    </article>
  </div>
</template>
