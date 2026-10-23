<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  donate,
  resetWheel,
  spinWheel,
  wheelSnapshot,
} from "../services/wheelService";
import type { WheelGame, WheelState } from "../types/wheel";

const palette = [
  "#b5bd68",
  "#7fb2d9",
  "#d98f7f",
  "#9d8fd9",
  "#7fd9b8",
  "#d9c27f",
];

const radius = 100;
const center = 110;

// Согласовано с transition в lab.css.
const spinTurns = 4;
const spinMs = 1800;

// Не совпадёт с именем игры: выбор сверяется по этому значению, а не по строке.
const newGameChoice = "＋";

const state = ref<WheelState>({ games: [], total: 0, spins: 0 });
const gameChoice = ref("");
const customName = ref("");
const amount = ref<number | null>(null);
const winner = ref("");
const spinning = ref(false);
const error = ref("");
const rotation = ref(0);
const rotor = ref<SVGGElement | null>(null);

const reducedMotion =
  typeof matchMedia === "function" &&
  matchMedia("(prefers-reduced-motion: reduce)").matches;

const money = new Intl.NumberFormat("be-BY", {
  style: "currency",
  currency: "BYN",
  currencyDisplay: "code",
  minimumFractionDigits: 2,
});
const integer = new Intl.NumberFormat("ru-RU");

const games = computed<WheelGame[]>(() => state.value.games);
const isEmpty = computed(() => games.value.length === 0);
const customNameNeeded = computed(
  () => isEmpty.value || gameChoice.value === newGameChoice,
);

const sectors = computed(() => {
  let angle = -Math.PI / 2; // отсчёт от 12 часов
  return games.value.map((game, index) => {
    const sector = {
      game,
      start: angle,
      sweep: game.probability * Math.PI * 2,
      color: palette[index % palette.length],
    };
    angle += sector.sweep;
    return sector;
  });
});

// Ломаная вместо дуги: у сектора с долей около единицы концы дуги почти
// совпадают, и WebKitGTK уводит такую дугу за viewBox — колесо исчезает.
function sectorPath(sector: { start: number; sweep: number }): string {
  const steps = Math.max(1, Math.ceil((sector.sweep * 180) / Math.PI));
  const point = (angle: number) =>
    `${(center + radius * Math.cos(angle)).toFixed(2)} ${(center + radius * Math.sin(angle)).toFixed(2)}`;

  let path = `M ${center} ${center} L ${point(sector.start)}`;
  for (let i = 1; i <= steps; i++) {
    path += ` L ${point(sector.start + (sector.sweep * i) / steps)}`;
  }

  return `${path} Z`;
}

function failure(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}

// Ведёт к середине сектора победителя и всегда вперёд, иначе анимация шла бы
// назад.
function turnTo(name: string) {
  const index = games.value.findIndex((game) => game.name === name);
  if (index < 0) return;

  const before = games.value
    .slice(0, index)
    .reduce((sum, game) => sum + game.probability, 0);
  const middle = (before + games.value[index].probability / 2) * 360;
  const current = rotation.value;
  const delta = (((-middle - current) % 360) + 360) % 360;

  rotation.value = current + delta + (reducedMotion ? 0 : spinTurns * 360);
}

function spinEnded(): Promise<void> {
  const element = rotor.value;
  if (!element || reducedMotion) return Promise.resolve();

  return new Promise((resolve) => {
    const finish = () => {
      element.removeEventListener("transitionend", onEnd);
      clearTimeout(timer);
      resolve();
    };
    const onEnd = (event: TransitionEvent) => {
      if (event.target === element && event.propertyName === "transform")
        finish();
    };
    // Если transition не сработает (скрытая вкладка), результат всё равно нужен.
    const timer = setTimeout(finish, spinMs + 400);
    element.addEventListener("transitionend", onEnd);
  });
}

async function submitDonation() {
  error.value = "";
  const name = customNameNeeded.value
    ? customName.value.trim()
    : gameChoice.value;
  const sum = Number(amount.value);
  if (
    name === "" ||
    name === newGameChoice ||
    !Number.isFinite(sum) ||
    sum <= 0
  ) {
    error.value = "Выберите игру и укажите сумму больше нуля.";
    return;
  }

  try {
    state.value = await donate(name, sum);
    gameChoice.value = "";
    customName.value = "";
    amount.value = null;
  } catch (err) {
    error.value = failure(err);
  }
}

async function spin() {
  error.value = "";
  spinning.value = true;
  winner.value = "";
  try {
    const result = await spinWheel();
    state.value = result.state;
    turnTo(result.winner);
    await spinEnded(); // подсветка — после остановки, а не на лету
    winner.value = result.winner;
  } catch (err) {
    error.value = failure(err);
  } finally {
    spinning.value = false;
  }
}

async function reset() {
  error.value = "";
  try {
    state.value = await resetWheel();
    winner.value = "";
    gameChoice.value = "";
    customName.value = "";
  } catch (err) {
    error.value = failure(err);
  }
}

onMounted(async () => {
  try {
    state.value = await wheelSnapshot();
  } catch (err) {
    error.value = failure(err);
  }
});
</script>

<template>
  <section class="panel">
    <div class="wheel-layout">
      <div class="wheel-figure">
        <svg
          viewBox="0 0 220 220"
          class="wheel-disc"
          role="img"
          :aria-label="
            isEmpty ? 'Колесо без секторов' : 'Колесо с секторами игр'
          "
        >
          <g
            ref="rotor"
            class="wheel-rotor"
            :style="{ transform: `rotate(${rotation}deg)` }"
          >
            <path
              v-for="sector in sectors"
              :key="sector.game.name"
              :d="sectorPath(sector)"
              :fill="sector.color"
              :class="['wheel-sector', { winner: sector.game.name === winner }]"
            />
          </g>
          <circle :cx="center" :cy="center" :r="16" class="wheel-hub" />
          <!-- Последним: иначе вращающиеся сектора перекрывают указатель. -->
          <polygon class="wheel-pointer" points="110,34 94,4 126,4" />
        </svg>

        <p v-if="isEmpty" class="wheel-caption">
          Колесо пусто: внесите хотя бы одно пожертвование.
        </p>
        <p v-else class="wheel-caption">
          Последнее вращение:
          <strong>{{ winner || "ещё не крутили" }}</strong>
        </p>
      </div>

      <div class="wheel-side">
        <!-- Валидация — в обработчике: нативная подсказка WebKitGTK бесполезна. -->
        <form class="category-row" novalidate @submit.prevent="submitDonation">
          <label v-if="!isEmpty && !customNameNeeded" class="field">
            <span>Игра</span>
            <select v-model="gameChoice">
              <option value="">Выберите игру</option>
              <option v-for="game in games" :key="game.name" :value="game.name">
                {{ game.name }}
              </option>
              <option :value="newGameChoice">Другая игра…</option>
            </select>
          </label>
          <label v-else class="field">
            <span>Игра</span>
            <input
              v-model="customName"
              type="text"
              placeholder="Witcher 3"
              autocomplete="off"
            />
          </label>
          <label class="field">
            <span>Сумма, BYN</span>
            <input
              v-model.number="amount"
              type="number"
              min="0.01"
              step="0.01"
            />
          </label>
          <!-- Пока колесо крутится, новые суммы меняли бы уже показанные вероятности. -->
          <button class="btn" type="submit" :disabled="spinning">
            Пожертвовать
          </button>
        </form>

        <div class="category-actions">
          <button
            class="btn"
            type="button"
            :disabled="spinning || isEmpty"
            @click="spin"
          >
            {{ spinning ? "Вращение..." : "Крутить колесо" }}
          </button>
          <button
            class="btn-small"
            type="button"
            :disabled="spinning || isEmpty"
            @click="reset"
          >
            Очистить
          </button>
          <span class="wheel-meta"
            >Всего: {{ money.format(state.total) }}</span
          >
        </div>

        <p v-if="error" class="error">{{ error }}</p>
      </div>
    </div>

    <article v-if="!isEmpty" class="result-card">
      <h3>Игры на колесе</h3>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Игра</th>
              <th>Пожертвования</th>
              <th>Вероятность</th>
              <th>Выпало</th>
              <th>Частота</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="game in games"
              :key="game.name"
              :class="{ 'wheel-row-winner': game.name === winner }"
            >
              <td>{{ game.name }}</td>
              <td>{{ money.format(game.amount) }}</td>
              <td>{{ game.probability.toFixed(4) }}</td>
              <td>{{ integer.format(game.wins) }}</td>
              <td>
                {{
                  state.spins > 0 ? (game.wins / state.spins).toFixed(4) : "—"
                }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </article>
  </section>
</template>
