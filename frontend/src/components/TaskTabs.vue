<script setup lang="ts">
import type { LabTask } from "../types/navigation";

defineProps<{
  tasks: LabTask[];
  active: string;
}>();

const emit = defineEmits<{ (event: "change", value: string): void }>();
</script>

<template>
  <nav class="task-tabs" aria-label="Задачи раздела">
    <button
      v-for="task in tasks"
      :key="task.id"
      :class="['task-tab', { active: task.id === active }]"
      type="button"
      :aria-current="task.id === active ? 'page' : undefined"
      @click="emit('change', task.id)"
    >
      {{ task.label }}
    </button>
  </nav>
</template>
