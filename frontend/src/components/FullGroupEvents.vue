<script setup lang="ts">
import { computed, ref } from "vue";
import EventResultView from "./EventResultView.vue";
import { SimulationResult } from "../types/randomEvents";
import { simulateFullGroupEvents } from "../services/randomEventsService";

const probabilities = ref<number[]>([0.2, 0.3, 0.5]);
const sampleSize = ref(1000000);
const result = ref<SimulationResult | null>(null);
const loading = ref(false);
const error = ref("");

const probabilitySum = computed(() =>
  probabilities.value.reduce((sum, p) => sum + p, 0),
);
const sumValid = computed(() => Math.abs(probabilitySum.value - 1) < 1e-6);
const valid = computed(
  () =>
    probabilities.value.length >= 2 &&
    probabilities.value.every((p) => p >= 0 && p <= 1) &&
    sampleSize.value >= 1,
);

function addProbability() {
  probabilities.value.push(0.1);
}

function removeProbability(index: number) {
  if (probabilities.value.length > 2) probabilities.value.splice(index, 1);
}

function normalize() {
  const sum = probabilitySum.value;
  if (sum <= 0) return;
  // Округляем до 6 знаков, а остаток отдаём наибольшей, чтобы сумма стала ровно 1.
  const scale = 1e6;
  const scaled = probabilities.value.map((p) => Math.round((p / sum) * scale));
  const largest = scaled.indexOf(Math.max(...scaled));
  scaled[largest] += scale - scaled.reduce((total, value) => total + value, 0);
  probabilities.value = scaled.map((value) => value / scale);
}

async function run() {
  error.value = "";
  if (!valid.value) {
    error.value = "Проверьте вероятности и N.";
    return;
  }
  if (!sumValid.value) {
    error.value = "Сумма вероятностей полной группы должна быть равна 1.";
    return;
  }
  loading.value = true;
  try {
    result.value = await simulateFullGroupEvents(
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
      <h2>Задание 4. Полная группа событий</h2>
      <p>Список вероятностей, сумма которых равна 1</p>
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
      <div class="category-actions">
        <button class="btn-small" type="button" @click="addProbability">
          Добавить
        </button>
        <button class="btn-small" type="button" @click="normalize">
          Нормировать
        </button>
        <span :class="['probability-sum', sumValid ? 'ok' : 'bad']"
          >Σ<var>p</var> = {{ probabilitySum.toFixed(4) }}</span
        >
      </div>
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
