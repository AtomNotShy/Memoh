<script setup lang="ts">
import { RouterView } from 'vue-router'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  Toaster,
} from '@memoh/ui'
import SvgIcon from '@jamescoyle/vue-icon'
import { mdiTranslate, mdiBrightness6 } from '@mdi/js'

// @ts-ignore
import 'vue-sonner/style.css'


</script>

<template>
  <section class="[&_input]:shadow-none!">
    <RouterView />
    <Toaster position="top-center" />
  </section>
</template>