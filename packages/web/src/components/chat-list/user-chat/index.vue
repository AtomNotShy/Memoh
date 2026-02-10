<template>
  <div class="flex">
    <p
      class="leading-7 not-first:mt-6 max-w-[90%] ml-auto text-muted-foreground bg-[#F9F9F9] p-4 rounded-xl rounded-tr-none break-all dark:bg-[#1C1917]
"
    >
      {{ userSay.description }}
    </p>
  </div>
</template>

<script setup lang="ts">
import type { user } from '@memoh/shared'
const { userSay } = defineProps<{
  userSay: user
}>()
</script>