<template>
  <section>
    <MainLayout>
      <template #sidebar>
        <SideBar />
      </template>
      <template #main>
        <MainContainer />
      </template>
    </MainLayout>
  </section>
</template>

<script setup lang="ts">
import MainLayout from '@/layout/mainLayout/index.vue'
import SideBar from '@/components/Sidebar/index.vue'
import MainContainer from '@/components/MainContainer/index.vue'
import { provide,ref } from 'vue'


provide('sideBarIsOpen',ref(true))
</script>