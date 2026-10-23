<script setup lang="ts">
import { computed, ref } from "vue";
import SectionTabs from "./components/SectionTabs.vue";
import TaskTabs from "./components/TaskTabs.vue";
import { sections } from "./sections";
import type { LabTask, SectionId } from "./types/navigation";

const activeSectionId = ref<SectionId>("events");

const activeTaskBySection = ref<Record<SectionId, string>>(
  Object.fromEntries(
    sections.map((section) => [section.id, section.tasks[0].id]),
  ) as Record<SectionId, string>,
);

const activeSection = computed(
  () =>
    sections.find((section) => section.id === activeSectionId.value) ??
    sections[0],
);
const activeTaskId = computed(
  () => activeTaskBySection.value[activeSectionId.value],
);
const activeTask = computed<LabTask>(
  () =>
    activeSection.value.tasks.find((task) => task.id === activeTaskId.value) ??
    activeSection.value.tasks[0],
);

function selectSection(id: SectionId) {
  activeSectionId.value = id;
}

function selectTask(id: string) {
  activeTaskBySection.value[activeSectionId.value] = id;
}
</script>

<template>
  <div class="app-shell">
    <header class="app-header">
      <!-- <div class="app-brand"> -->
      <!--   <span class="app-brand-mark">S</span> -->
      <!--   <span class="app-brand-name">Simulation Suite</span> -->
      <!-- </div> -->
      <SectionTabs
        :sections="sections"
        :active="activeSectionId"
        @change="selectSection"
      />
    </header>

    <TaskTabs
      :tasks="activeSection.tasks"
      :active="activeTaskId"
      @change="selectTask"
    />

    <main class="content">
      <!-- KeepAlive: у каждой задачи свой кэш, поэтому введённые данные и
           результаты не сбрасываются при переключении разделов и задач. -->
      <KeepAlive>
        <component
          :is="activeTask.component"
          :key="`${activeSection.id}:${activeTask.id}`"
          v-bind="activeTask.props ?? {}"
        />
      </KeepAlive>
    </main>
  </div>
</template>
