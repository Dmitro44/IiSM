<script setup lang="ts">
defineProps<{
  title: string;
  description: string;
}>();
</script>

<template>
  <section class="panel">
    <header class="panel-header placeholder-header">
      <div>
        <h2>{{ title }}</h2>
        <p>{{ description }}</p>
      </div>
      <span class="status-chip">В разработке</span>
    </header>

    <div class="placeholder">
      <h3>Экран ещё не реализован</h3>
      <p>
        Навигация уже готова: раздел и задача переключаются, состояние вкладок
        сохраняется. Расчёты и графики появятся на следующем шаге.
      </p>
    </div>
  </section>
</template>
