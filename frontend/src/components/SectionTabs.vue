<script setup lang="ts">
import type { LabSection, SectionId } from "../types/navigation";
import SectionIcon from "./SectionIcon.vue";

defineProps<{
  sections: LabSection[];
  active: SectionId;
}>();

const emit = defineEmits<{ (event: "change", value: SectionId): void }>();
</script>

<template>
  <nav class="section-tabs" aria-label="Разделы">
    <button
      v-for="section in sections"
      :key="section.id"
      :class="['section-tab', { active: section.id === active }]"
      type="button"
      :aria-current="section.id === active ? 'page' : undefined"
      @click="emit('change', section.id)"
    >
      <SectionIcon :name="section.icon" />
      <span>{{ section.label }}</span>
    </button>
  </nav>
</template>
