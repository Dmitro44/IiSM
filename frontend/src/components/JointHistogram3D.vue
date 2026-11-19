<script setup lang="ts">
import { computed, ref } from "vue";
import type { JointCellPoint } from "../types/bivariateUi";
import { formatNumber } from "../utils/format";

const props = withDefaults(
  defineProps<{
    cells: JointCellPoint[];
    gridSize: number;
    title?: string;
  }>(),
  {
    title: "Совместное распределение",
  },
);

const VIEW_W = 620;
const VIEW_H = 430;
const PADDING = 6;
const MAX_BAR_HEIGHT = 1.2;

// Камера фиксирована: масштаб и центр считаются один раз по начальному
// ракурсу, поэтому при вращении график не отъезжает и не приезжает.
const FIT_MARGIN = 1.0;
const DEFAULT_YAW = Math.PI / 4;
const DEFAULT_PITCH = Math.PI / 6;
const MIN_PITCH = 0.2;
const MAX_PITCH = Math.PI / 2 - 0.08;

const yaw = ref(DEFAULT_YAW);
const pitch = ref(DEFAULT_PITCH);
const dragging = ref(false);

const dragStart = {
  x: 0,
  y: 0,
  yaw: DEFAULT_YAW,
  pitch: DEFAULT_PITCH,
  pointerId: -1,
};

interface Vec3 {
  x: number;
  y: number;
  z: number;
}

interface ScreenPoint {
  x: number;
  y: number;
}

interface CellValue {
  index: number;
  i: number;
  j: number;
  count: number;
  expected: number;
  centerX: number;
  centerY: number;
}

interface BarFace {
  points: Vec3[];
  className: string;
  depth: number;
  visible: boolean;
}

interface BarModel {
  key: number;
  depth: number;
  faces: BarFace[];
  top: Vec3[];
  title: string;
}

interface ScreenFace {
  key: string;
  points: string;
  className: string;
}

interface ScreenBar {
  key: number;
  title: string;
  faces: ScreenFace[];
}

interface Scene {
  bars: ScreenBar[];
  meshLines: string[];
  meshDots: ScreenPoint[];
  floor: string;
}

const emptyScene: Scene = {
  bars: [],
  meshLines: [],
  meshDots: [],
  floor: "",
};

const gridSizeSafe = computed(() =>
  Number.isFinite(props.gridSize) ? Math.max(0, Math.floor(props.gridSize)) : 0,
);

const cellCount = computed(() => gridSizeSafe.value * gridSizeSafe.value);

const cellsMismatch = computed(
  () => gridSizeSafe.value > 0 && props.cells.length !== cellCount.value,
);

const cellData = computed<CellValue[]>(() => {
  const n = gridSizeSafe.value;
  const result: CellValue[] = [];
  if (n <= 0) return result;

  for (let i = 0; i < n; i++) {
    for (let j = 0; j < n; j++) {
      const index = i * n + j;
      const cell: JointCellPoint | undefined = props.cells[index];
      if (!cell) continue;
      result.push({
        index,
        i,
        j,
        count: Number.isFinite(cell.count) ? Math.max(cell.count, 0) : 0,
        expected: Number.isFinite(cell.expectedCount) ? Math.max(cell.expectedCount, 0) : 0,
        centerX: Number.isFinite(cell.centerX) ? cell.centerX : 0,
        centerY: Number.isFinite(cell.centerY) ? cell.centerY : 0,
      });
    }
  }
  return result;
});

const maxObserved = computed(() =>
  cellData.value.reduce((max, cell) => Math.max(max, cell.count), 0),
);

const maxExpected = computed(() =>
  cellData.value.reduce((max, cell) => Math.max(max, cell.expected), 0),
);

const heightScale = computed(() => {
  if (maxObserved.value > 0) return MAX_BAR_HEIGHT / maxObserved.value;
  if (maxExpected.value > 0) return MAX_BAR_HEIGHT / maxExpected.value;
  return 0;
});

const hasScene = computed(() => gridSizeSafe.value > 0 && props.cells.length > 0);

function clamp(value: number, lo: number, hi: number): number {
  return Math.min(Math.max(value, lo), hi);
}

// Ортографическая проекция: поворот вокруг вертикали (yaw), затем наклон
// камеры (pitch). Масштаб сюда не входит — он фиксирован ниже.
function projectWith(point: Vec3, yawValue: number, pitchValue: number): ScreenPoint {
  const center = gridSizeSafe.value / 2;
  const px = point.x - center;
  const py = point.y - center;
  const cosYaw = Math.cos(yawValue);
  const sinYaw = Math.sin(yawValue);
  return {
    x: px * cosYaw - py * sinYaw,
    y:
      (px * sinYaw + py * cosYaw) * Math.sin(pitchValue) -
      point.z * Math.cos(pitchValue),
  };
}

function project(point: Vec3): ScreenPoint {
  return projectWith(point, yaw.value, pitch.value);
}

function depthOf(point: Vec3): number {
  const center = gridSizeSafe.value / 2;
  return (
    (point.x - center) * Math.sin(yaw.value) +
    (point.y - center) * Math.cos(yaw.value)
  );
}

function polygon(points: Vec3[], toScreen: (p: Vec3) => ScreenPoint): string {
  return points
    .map((point) => {
      const screen = toScreen(point);
      return `${screen.x.toFixed(2)},${screen.y.toFixed(2)}`;
    })
    .join(" ");
}

function faceDepth(points: Vec3[]): number {
  return points.reduce((sum, point) => sum + depthOf(point), 0) / points.length;
}

// Модель строится в координатах сетки: ячейка (i, j) занимает квадрат
// [i, i+1] x [j, j+1], высота — count * heightScale. Рисуем все четыре
// боковые грани: иначе между соседними столбцами при повороте появляются
// щели, через которые видно фон.
const modelBars = computed<BarModel[]>(() => {
  const scale = heightScale.value;
  const sinYaw = Math.sin(yaw.value);
  const cosYaw = Math.cos(yaw.value);
  return cellData.value.map((cell) => {
    const x0 = cell.i;
    const x1 = cell.i + 1;
    const y0 = cell.j;
    const y1 = cell.j + 1;
    const h = cell.count * scale;

    // Видимые грани остаются полупрозрачными, как в исходном графике;
    // дальние грани рисуем непрозрачной тёмной подложкой — она закрывает
    // щели между столбцами и не просвечивает через ближнюю грань.
    const faces: BarFace[] = [
      {
        points: [
          { x: x0, y: y1, z: 0 },
          { x: x0, y: y0, z: 0 },
          { x: x0, y: y0, z: h },
          { x: x0, y: y1, z: h },
        ],
        className: `bar-face ${sinYaw <= 0 ? "bar-x" : "bar-back"}`,
        depth: 0,
        visible: sinYaw <= 0,
      },
      {
        points: [
          { x: x1, y: y0, z: 0 },
          { x: x1, y: y1, z: 0 },
          { x: x1, y: y1, z: h },
          { x: x1, y: y0, z: h },
        ],
        className: `bar-face ${sinYaw >= 0 ? "bar-x" : "bar-back"}`,
        depth: 0,
        visible: sinYaw >= 0,
      },
      {
        points: [
          { x: x0, y: y0, z: 0 },
          { x: x1, y: y0, z: 0 },
          { x: x1, y: y0, z: h },
          { x: x0, y: y0, z: h },
        ],
        className: `bar-face ${cosYaw <= 0 ? "bar-y" : "bar-back"}`,
        depth: 0,
        visible: cosYaw <= 0,
      },
      {
        points: [
          { x: x1, y: y1, z: 0 },
          { x: x0, y: y1, z: 0 },
          { x: x0, y: y1, z: h },
          { x: x1, y: y1, z: h },
        ],
        className: `bar-face ${cosYaw >= 0 ? "bar-y" : "bar-back"}`,
        depth: 0,
        visible: cosYaw >= 0,
      },
    ];
    for (const face of faces) {
      face.depth = faceDepth(face.points);
    }

    return {
      key: cell.index,
      depth: (x0 + 0.5) * Math.sin(yaw.value) + (y0 + 0.5) * Math.cos(yaw.value),
      faces,
      top: [
        { x: x0, y: y0, z: h },
        { x: x1, y: y0, z: h },
        { x: x1, y: y1, z: h },
        { x: x0, y: y1, z: h },
      ],
      title: `центр (${formatNumber(cell.centerX, 3)}; ${formatNumber(
        cell.centerY,
        3,
      )}): наблюдений ${formatNumber(cell.count, 0)}, ожидается ${formatNumber(
        cell.expected,
        2,
      )}`,
    };
  });
});

const scene = computed<Scene>(() => {
  const n = gridSizeSafe.value;
  const bars = modelBars.value;
  if (n <= 0 || bars.length === 0) return emptyScene;

  const scale = heightScale.value;
  const floor: Vec3[] = [
    { x: 0, y: 0, z: 0 },
    { x: n, y: 0, z: 0 },
    { x: n, y: n, z: 0 },
    { x: 0, y: n, z: 0 },
  ];

  const meshPoints: Array<{ i: number; j: number; point: Vec3 }> = [];
  for (const cell of cellData.value) {
    meshPoints.push({
      i: cell.i,
      j: cell.j,
      point: {
        x: cell.i + 0.5,
        y: cell.j + 0.5,
        z: cell.expected * scale,
      },
    });
  }

  // Границы и масштаб считаем по начальному ракурсу и больше не меняем.
  // Иначе камера «дышит» на каждом кадре вращения.
  const projected: ScreenPoint[] = [];
  const addReference = (point: Vec3) =>
    projected.push(projectWith(point, DEFAULT_YAW, DEFAULT_PITCH));
  for (const point of floor) addReference(point);
  for (const bar of bars) {
    for (const face of bar.faces) {
      for (const point of face.points) addReference(point);
    }
    for (const point of bar.top) addReference(point);
  }
  for (const entry of meshPoints) addReference(entry.point);

  let minX = Infinity;
  let maxX = -Infinity;
  let minY = Infinity;
  let maxY = -Infinity;
  for (const point of projected) {
    minX = Math.min(minX, point.x);
    maxX = Math.max(maxX, point.x);
    minY = Math.min(minY, point.y);
    maxY = Math.max(maxY, point.y);
  }

  const spanX = Math.max(maxX - minX, 0.0001);
  const spanY = Math.max(maxY - minY, 0.0001);
  const fitScale =
    Math.min((VIEW_W - 2 * PADDING) / spanX, (VIEW_H - 2 * PADDING) / spanY) *
    FIT_MARGIN;
  const offsetX = (VIEW_W - spanX * fitScale) / 2 - minX * fitScale;
  const offsetY = (VIEW_H - spanY * fitScale) / 2 - minY * fitScale;

  const toScreen = (point: Vec3): ScreenPoint => {
    const screen = project(point);
    return {
      x: screen.x * fitScale + offsetX,
      y: screen.y * fitScale + offsetY,
    };
  };

  const screenBars: ScreenBar[] = [...bars]
    .sort((a, b) => a.depth - b.depth || a.key - b.key)
    .map((bar) => ({
      key: bar.key,
      title: bar.title,
      faces: [
        ...[...bar.faces]
          .sort((a, b) => {
            if (a.visible !== b.visible) return a.visible ? 1 : -1;
            return a.depth - b.depth;
          })
          .map((face, index) => ({
            key: `${bar.key}-face-${index}`,
            points: polygon(face.points, toScreen),
            className: face.className,
          })),
        {
          key: `${bar.key}-top`,
          points: polygon(bar.top, toScreen),
          className: "bar-face bar-top",
        },
      ],
    }));

  const rowMap = new Map<number, Array<{ j: number; point: Vec3 }>>();
  const columnMap = new Map<number, Array<{ i: number; point: Vec3 }>>();
  for (const entry of meshPoints) {
    const row = rowMap.get(entry.i) ?? [];
    row.push({ j: entry.j, point: entry.point });
    rowMap.set(entry.i, row);

    const column = columnMap.get(entry.j) ?? [];
    column.push({ i: entry.i, point: entry.point });
    columnMap.set(entry.j, column);
  }

  const meshLines: string[] = [];
  for (const row of rowMap.values()) {
    row.sort((a, b) => a.j - b.j);
    if (row.length >= 2) {
      meshLines.push(polygon(row.map((entry) => entry.point), toScreen));
    }
  }
  for (const column of columnMap.values()) {
    column.sort((a, b) => a.i - b.i);
    if (column.length >= 2) {
      meshLines.push(polygon(column.map((entry) => entry.point), toScreen));
    }
  }

  return {
    bars: screenBars,
    meshLines,
    meshDots: meshPoints.map((entry) => toScreen(entry.point)),
    floor: polygon(floor, toScreen),
  };
});

function resetView() {
  yaw.value = DEFAULT_YAW;
  pitch.value = DEFAULT_PITCH;
}

function onPointerDown(event: PointerEvent) {
  if (event.button !== 0) return;
  dragging.value = true;
  dragStart.x = event.clientX;
  dragStart.y = event.clientY;
  dragStart.yaw = yaw.value;
  dragStart.pitch = pitch.value;
  dragStart.pointerId = event.pointerId;
  (event.currentTarget as SVGSVGElement).setPointerCapture(event.pointerId);
  event.preventDefault();
}

function onPointerMove(event: PointerEvent) {
  if (!dragging.value || event.pointerId !== dragStart.pointerId) return;
  yaw.value = dragStart.yaw - (event.clientX - dragStart.x) * 0.012;
  pitch.value = clamp(
    dragStart.pitch + (event.clientY - dragStart.y) * 0.008,
    MIN_PITCH,
    MAX_PITCH,
  );
}

function onPointerUp(event: PointerEvent) {
  if (!dragging.value) return;
  dragging.value = false;
  const element = event.currentTarget as SVGSVGElement;
  if (element.hasPointerCapture(event.pointerId)) {
    element.releasePointerCapture(event.pointerId);
  }
}
</script>

<template>
  <figure class="joint-chart">
    <figcaption class="joint-head">
      <h4>{{ title }}</h4>
      <div class="joint-legend">
        <span><i class="legend-bar"></i>наблюдения</span>
        <span><i class="legend-mesh"></i>теоретическая плотность</span>
      </div>
    </figcaption>

    <div class="joint-body">
      <svg
        :viewBox="`0 0 ${VIEW_W} ${VIEW_H}`"
        :class="{ dragging }"
        role="img"
        :aria-label="`${title}. Потяните мышью, чтобы повернуть.`"
        @pointerdown="onPointerDown"
        @pointermove="onPointerMove"
        @pointerup="onPointerUp"
        @pointercancel="onPointerUp"
        @dblclick="resetView"
      >
        <template v-if="hasScene">
          <polygon :points="scene.floor" class="floor" />

          <g v-for="bar in scene.bars" :key="bar.key" class="bar-group">
            <title>{{ bar.title }}</title>
            <polygon
              v-for="face in bar.faces"
              :key="face.key"
              :points="face.points"
              :class="face.className"
            />
          </g>

          <g class="mesh">
            <polyline
              v-for="(line, index) in scene.meshLines"
              :key="`mesh-${index}`"
              :points="line"
            />
            <circle
              v-for="(dot, index) in scene.meshDots"
              :key="`dot-${index}`"
              :cx="dot.x"
              :cy="dot.y"
              r="1.8"
            />
          </g>
        </template>
      </svg>
    </div>

    <div class="joint-foot">
      <p v-if="cellsMismatch" class="joint-warning">
        Данные не совпадают с сеткой: ожидалось {{ cellCount }} ячеек, получено
        {{ cells.length }} — показаны только доступные.
      </p>
      <p v-if="!hasScene" class="joint-note">Нет данных для построения совместной гистограммы.</p>
      <p v-else class="joint-note">
        Сетка {{ gridSizeSafe }}×{{ gridSizeSafe }} — разбиение диска на равные ячейки; высота
        столбца и узлы сетки нормированы по наблюдениям. Потяните мышью, чтобы
        повернуть.
      </p>
      <button class="btn-small" type="button" @click="resetView">Сбросить вид</button>
    </div>
  </figure>
</template>

<style scoped>
.joint-chart {
  margin: 0;
  padding: var(--s-2);
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--glass-border);
  border-radius: var(--radius);
}

.joint-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 0.5rem 1rem;
}

.joint-head h4 {
  margin: 0;
  color: var(--text);
  font-size: 1rem;
}

.joint-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 0.9rem;
  color: var(--muted);
  font-size: 0.78rem;
}

.joint-legend span {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}

.legend-bar {
  width: 0.7rem;
  height: 0.7rem;
  border-radius: 2px;
  background: var(--accent);
}

.legend-mesh {
  position: relative;
  width: 1.2rem;
  height: 1px;
  background: rgba(255, 255, 255, 0.55);
}

.legend-mesh::after {
  content: "";
  position: absolute;
  top: 50%;
  left: 50%;
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.8);
  transform: translate(-50%, -50%);
}

.joint-body {
  margin-top: var(--s-1);
}

svg {
  display: block;
  width: 100%;
  height: auto;
  overflow: visible;
  cursor: grab;
  touch-action: none;
}

svg.dragging {
  cursor: grabbing;
}

.floor {
  fill: rgba(255, 255, 255, 0.02);
  stroke: rgba(255, 255, 255, 0.09);
  stroke-width: 1;
}

.bar-face {
  stroke: rgba(20, 24, 21, 0.55);
  stroke-width: 0.8;
  stroke-linejoin: round;
}

.bar-back {
  fill: #2f3428;
}

.bar-top {
  fill: var(--accent);
  fill-opacity: 0.82;
}

.bar-x {
  fill: var(--accent);
  fill-opacity: 0.46;
}

.bar-y {
  fill: var(--accent);
  fill-opacity: 0.28;
}

.mesh {
  pointer-events: none;
}

.mesh polyline {
  fill: none;
  stroke: rgba(255, 255, 255, 0.55);
  stroke-width: 1.4;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.mesh circle {
  fill: rgba(255, 255, 255, 0.8);
}

.joint-foot {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--s-2);
  flex-wrap: wrap;
}

.joint-note,
.joint-warning {
  flex: 1 1 24rem;
  margin: var(--s-1) 0 0;
  color: var(--muted);
  font-size: 0.78rem;
  line-height: 1.45;
}

.joint-warning {
  color: #ffb3bf;
}

.joint-foot .btn-small {
  flex: 0 0 auto;
  margin-top: var(--s-1);
}
</style>
