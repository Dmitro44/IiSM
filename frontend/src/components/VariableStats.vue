<script setup lang="ts">
import type {
  GoodnessOfFit,
  Interval,
  SampleStats,
} from "../../bindings/github.com/Dmitro44/IiSM/internal/randomvars/models";
import { formatCount, formatNumber, formatPValue } from "../utils/format";

defineProps<{
  sampleSize: number;
  stats: SampleStats;
  theoreticalMean: number;
  theoreticalVariance: number;
  meanCi: Interval;
  varianceCi: Interval;
  test: GoodnessOfFit;
}>();
</script>

<template>
  <article class="result-card">
    <h3>Точечные оценки</h3>
    <div class="stats-grid">
      <div>
        <strong>Среднее <var>x̄</var></strong>
        <span>{{ formatNumber(stats.mean) }}</span>
        <!-- <small>теория: {{ formatNumber(theoreticalMean) }}</small> -->
      </div>
      <div>
        <strong>Дисперсия <var>s²</var></strong>
        <span>{{ formatNumber(stats.variance) }}</span>
        <!-- <small>теория: {{ formatNumber(theoreticalVariance) }}</small> -->
      </div>
      <div>
        <strong>Ср. квадратичное <var>s</var></strong>
        <span>{{ formatNumber(stats.stdDev) }}</span>
      </div>
      <!-- <div>
        <strong>Минимум</strong>
        <span>{{ formatNumber(stats.min) }}</span>
      </div>
      <div>
        <strong>Максимум</strong>
        <span>{{ formatNumber(stats.max) }}</span>
      </div> -->
      <div>
        <strong>Объём выборки</strong>
        <span>{{ formatCount(sampleSize) }}</span>
      </div>
    </div>

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
          <tr>
            <td>Математическое ожидание</td>
            <td>{{ formatNumber(meanCi.lower) }}</td>
            <td>{{ formatNumber(meanCi.upper) }}</td>
            <td>{{ formatNumber(meanCi.confidence, 3) }}</td>
          </tr>
          <tr>
            <td>Дисперсия</td>
            <td>{{ formatNumber(varianceCi.lower) }}</td>
            <td>{{ formatNumber(varianceCi.upper) }}</td>
            <td>{{ formatNumber(varianceCi.confidence, 3) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <!-- <p class="result-note">
      Для среднего использован интервал Стьюдента, для дисперсии — интервал
      через распределение <var>χ²</var>. <var>χ²</var>-интервал дисперсии
      точен для нормального закона; для остальных распределений это приближение
      при больших N.
    </p> -->

    <h3>Проверка гипотезы</h3>
    <div v-if="test.valid" class="test-card">
      <div class="test-head">
        <strong>{{ test.name }}</strong>
        <span
          :class="['test-verdict', test.accepted ? 'accepted' : 'rejected']"
        >
          {{
            test.accepted ? "Гипотеза не отвергается" : "Гипотеза отвергается"
          }}
        </span>
      </div>
      <div class="stats-grid test-stats">
        <div>
          <strong>Статистика <var>χ²</var></strong>
          <span>{{ formatNumber(test.statistic) }}</span>
        </div>
        <div>
          <strong>Степени свободы</strong>
          <span>{{ test.degreesOfFreedom }}</span>
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
      <p class="result-note">
        Если <var>p</var> &ge; <var>α</var>, расхождение выборки с заданным
        законом считается случайным. Иначе гипотеза о законе отвергается.
      </p>
    </div>
    <div v-else class="test-card">
      <strong>{{ test.name }}</strong>
      <p class="result-note">
        После объединения интервалов с ожидаемой частотой меньше 5 осталась одна
        группа, поэтому критерий <var>χ²</var> применить нельзя. Увеличьте N или
        число значений.
      </p>
    </div>
  </article>
</template>
