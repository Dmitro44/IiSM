<script setup lang="ts">
import { computed, ref } from "vue";
import EventResultView from "./EventResultView.vue";
import { SimulationResult } from "../types/randomEvents";
import { simulateComplexEvents } from "../services/randomEventsService";

const MAX_EVENTS = 10;

const probabilities = ref<number[]>([0.3, 0.5, 0.7]);
const sampleSize = ref(1000000);
const result = ref<SimulationResult | null>(null);
const loading = ref(false);
const error = ref("");

const valid = computed(
  () =>
    probabilities.value.length > 0 &&
    probabilities.value.length <= MAX_EVENTS &&
    probabilities.value.every((p) => p >= 0 && p <= 1) &&
    sampleSize.value >= 1,
);

const outcomesCount = computed(() => 1 << probabilities.value.length);

function addProbability() {
  probabilities.value.push(0.5);
}

function removeProbability(index: number) {
  if (probabilities.value.length > 1) probabilities.value.splice(index, 1);
}

async function run() {
  error.value = "";
  if (!valid.value) {
    error.value = `Проверьте вероятности и N: событий должно быть не больше ${MAX_EVENTS}.`;
    return;
  }
  loading.value = true;
  try {
    result.value = await simulateComplexEvents(
      probabilities.value,
      sampleSize.value,
    );
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err);
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <section class="panel">
    <header class="panel-header">
      <h2>Задание 2. Независимые события</h2>
      <p>
        A₁…Aₖ — независимые события. В каждом испытании выпадает ровно один из
        {{ outcomesCount }} совместных исходов.
      </p>
    </header>

    <div class="category-editor">
      <div class="prob-grid">
        <div v-for="(p, index) in probabilities" :key="index" class="prob-item">
          <div class="prob-name">A{{ index + 1 }}</div>
          <input
            v-model.number="probabilities[index]"
            type="number"
            min="0"
            max="1"
            step="0.05"
          />
          <button
            class="btn-small"
            type="button"
            @click="removeProbability(index)"
          >
            Удалить
          </button>
        </div>
      </div>
      <button class="btn-small" type="button" @click="addProbability">
        Добавить вероятность
      </button>
    </div>

    <div class="controls-grid">
      <label class="field">
        <span>N</span>
        <input
          v-model.number="sampleSize"
          type="number"
          min="1"
          max="1000000"
        />
      </label>
    </div>

    <div class="actions">
      <button class="btn" :disabled="loading" @click="run">
        {{ loading ? "Расчёт..." : "Выполнить" }}
      </button>
    </div>

    <p v-if="error" class="error">{{ error }}</p>
    <EventResultView :result="result" />
  </section>
</template>
