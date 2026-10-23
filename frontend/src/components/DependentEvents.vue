<script setup lang="ts">
import { computed, ref } from "vue";
import EventResultView from "./EventResultView.vue";
import { SimulationResult } from "../types/randomEvents";
import { simulateDependentEvents } from "../services/randomEventsService";

const probabilityA = ref(0.4);
const conditionalProbabilityB = ref(0.7);
const sampleSize = ref(1000000);
const result = ref<SimulationResult | null>(null);
const loading = ref(false);
const error = ref("");

const valid = computed(
  () =>
    probabilityA.value >= 0 &&
    probabilityA.value <= 1 &&
    conditionalProbabilityB.value >= 0 &&
    conditionalProbabilityB.value <= 1 &&
    sampleSize.value >= 1
);

async function run() {
  error.value = "";
  if (!valid.value) {
    error.value = "Проверьте P(A), P(B|A) и N.";
    return;
  }
  loading.value = true;
  try {
    result.value = await simulateDependentEvents(
      probabilityA.value,
      conditionalProbabilityB.value,
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
      <h2>Задание 3. Зависимые события</h2>
      <p>Симуляция AB, A¬B, ¬AB, ¬A¬B с P(A̅)=1−P(A), P(B|A̅)=1−P(B|A).</p>
    </header>

    <div class="controls-grid">
      <label class="field">
        <span>P(A)</span>
        <input v-model.number="probabilityA" type="number" min="0" max="1" step="0.05" />
      </label>
      <label class="field">
        <span>P(B|A)</span>
        <input v-model.number="conditionalProbabilityB" type="number" min="0" max="1" step="0.05" />
      </label>
      <label class="field">
        <span>N</span>
        <input v-model.number="sampleSize" type="number" min="1" max="1000000" />
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
