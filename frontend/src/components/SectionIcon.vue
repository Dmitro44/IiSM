<script setup lang="ts">
import type { SectionIconName } from "../types/navigation";

defineProps<{ name: SectionIconName }>();
</script>

<template>
  <svg
    v-if="name === 'events'"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.6"
    aria-hidden="true"
  >
    <circle cx="9" cy="9" r="5" />
    <circle cx="15" cy="9" r="5" />
    <circle cx="12" cy="15" r="5" />
  </svg>

  <svg
    v-else-if="name === 'variables'"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.6"
    stroke-linecap="round"
    aria-hidden="true"
  >
    <path d="M3 19c4.2-13.5 13.8-13.5 18 0" />
    <path d="M3 19h18" />
  </svg>

  <svg
    v-else-if="name === 'bivariate'"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.6"
    stroke-linecap="round"
    aria-hidden="true"
  >
    <path d="M4 3v17h17" />
    <circle cx="9" cy="15" r="1.3" />
    <circle cx="13" cy="10" r="1.3" />
    <circle cx="17" cy="13" r="1.3" />
    <circle cx="19" cy="7" r="1.3" />
  </svg>

  <svg
    v-else-if="name === 'queueing'"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.6"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <path d="M3 7h10M3 12h13M3 17h7" />
    <path d="M17 7l3 3-3 3" />
  </svg>

  <svg
    v-else
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.6"
    aria-hidden="true"
  >
    <circle cx="6" cy="6" r="2.5" />
    <circle cx="18" cy="10" r="2.5" />
    <circle cx="8" cy="18" r="2.5" />
    <path d="M8.5 6.8l7 2.4M16 12.2l-6.2 4.2" stroke-linecap="round" />
  </svg>
</template>
