<template>
  <aside class="[&_[data-state=collapsed]_:is(.title-container,.exist-btn)]:hidden">
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <div class="flex flex-row items-center w-full gap-2 px-3 py-2">
              <img
                src="/logo.png"
                class="size-10"
                alt="logo.png"
              >
              <span class="text-xl font-bold text-gray-500 dark:text-gray-400">
                Memoh
              </span>
            </div>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent class="[&_ul+ul]:mt-2!">
            <SidebarMenu
              v-for="sidebarItem in sidebarInfo"
              :key="sidebarItem.title"
            >
              <SidebarMenuItem class="[&_[aria-pressed=true]]:bg-accent!">
                <SidebarMenuButton
                  as-child
                  class="justify-start py-5! px-4"
                  :tooltip="sidebarItem.title"
                >
                  <Toggle
                    :class="` border border-transparent w-full flex justify-start ${curSlider === sidebarItem.name ? 'border-inherit' : ''}`"
                    :model-value="curSelectSlide(sidebarItem.name as string).value"
                    @update:model-value="(isSelect) => {
                      if (isSelect) {
                        curSlider = sidebarItem.name
                      }
                    }"
                    @click="router.push({ name: sidebarItem.name })"
                  >
                    <FontAwesomeIcon :icon="sidebarItem.icon" />
                    <span>{{ sidebarItem.title }}</span>
                  </Toggle>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
    </Sidebar>
  </aside>
</template>
<script setup lang="ts">
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  
  Toggle
} from '@memoh/ui'
import { computed } from 'vue'
import { useRouter,useRoute } from 'vue-router'
import { useUserStore } from '@/store/user'
import i18n from '@/i18n'
import { ref } from 'vue'


const router = useRouter()
const route=useRoute()

const { t } = i18n.global
const curSlider = ref()
const curSelectSlide = (cur: string) => computed(() => {
  return curSlider.value === cur||new RegExp(`^/main/${cur}$`).test(route.path)
})
const sidebarInfo = computed(() => [
  {
    title: t('slidebar.chat'),
    name: 'chat',
    icon: ['far', 'comments']
  },
  // {
  //   title: t('slidebar.home'),
  //   name: 'home',
  //   icon: mdiHome
  // },
  {
    title: t('slidebar.model_setting'),
    name: 'models',
    icon: ['fas', 'robot']
  }, {
    title: t('slidebar.setting'),
    name: 'settings',
    icon: ['fas', 'gear']
  },
  // {
  //   title: 'MCP',
  //   name: 'mcp',
  //   icon: mdiListBox
  // }, {
  //   title: t('slidebar.platform'),
  //   name: 'platform',
  //   icon: mdiBookArrowDown
  // }
])

</script>