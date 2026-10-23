<script setup lang="ts">
import { ref } from "vue";
import EventResultView from "./EventResultView.vue";
import { SimulationResult } from "../types/randomEvents";
import { simulateSimpleEvent } from "../services/randomEventsService";

const probability = ref(0.5);
const sampleSize = ref(1000000);
const result = ref<SimulationResult | null>(null);
const loading = ref(false);
const error = ref("");

async function run() {
  error.value = "";
  if (probability.value < 0 || probability.value > 1) {
    error.value = "Вероятность должна быть в интервале [0, 1].";
    return;
  }
  loading.value = true;
  try {
    result.value = await simulateSimpleEvent(probability.value, sampleSize.value);
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
      <h2>Задание 1. Простое случайное событие</h2>
      <p>Вероятность P(A) задаёт возможность происшествия события A.</p>
    </header>

    <div class="controls-grid">
      <label class="field">
        <span>P(A)</span>
        <input
          v-model.number="probability"
          type="number"
          min="0"
          max="1"
          step="0.05"
        />
      </label>
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
