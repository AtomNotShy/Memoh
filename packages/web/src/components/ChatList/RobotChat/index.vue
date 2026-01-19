<template>
  <div class="flex gap-4 items-start">
    <div class=" p-2 rounded-full bg-[#F9F9F9]">
      <svg-icon
        type="mdi"
        :path="mdiRobotOutline"
      />
    </div>
    <section>
      <sup class="font-semibold">
        {{ robotSay.type }}
      </sup>
      <p class="leading-7 text-muted-foreground">
        {{ robotSay.description }}
      </p>
    </section>
  </div>
</template>

<script setup lang="ts">
import SvgIcon from '@jamescoyle/vue-icon'
import { mdiRobotOutline } from '@mdi/js'
import type {robot}  from '@memoh/shared'


const {robotSay}=defineProps<{
  robotSay: robot
}>()
</script>