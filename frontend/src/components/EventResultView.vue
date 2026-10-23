<script setup lang="ts">
import { SimulationResult } from "../types/randomEvents";

defineProps<{
  result: SimulationResult | null;
}>();
</script>

<template>
  <article v-if="result" class="result-card">
    <h3>{{ result.title }}</h3>
    <p>N = {{ result.sampleSize.toLocaleString('ru-RU') }}</p>

    <div class="table-wrap">
      <table>
        <thead>
        <tr>
          <th>Событие / исход</th>
          <th>Число исходов</th>
          <th>Наблюдаемая частота</th>
          <th>Теоретическая вероятность</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="outcome in result.outcomes" :key="outcome.label">
          <td>{{ outcome.label }}</td>
          <td>{{ outcome.count.toLocaleString('ru-RU') }}</td>
          <td>{{ outcome.frequency.toFixed(6) }}</td>
          <td>{{ outcome.probability.toFixed(6) }}</td>
        </tr>
        </tbody>
      </table>
    </div>
  </article>
</template>
