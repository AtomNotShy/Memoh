<template>
  <aside class="[&_[data-state=collapsed]_:is(.title-container,.exist-btn)]:hidden">
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <button
          class="w-full flex items-center gap-2 px-3 py-2 text-left"
          @click="onLogoClick"
        >
          <img
            src="/logo.png"
            class="size-8 shrink-0"
            alt="logo"
          >
          <span class="text-xl font-bold text-gray-500 dark:text-gray-400 group-data-[state=collapsed]:hidden">
            Memoh
          </span>
        </button>
      </SidebarHeader>

      <SidebarContent>
        <component
          :is="currentListComponent"
          :collapsible="true"
        />
      </SidebarContent>
      <SidebarFooter class="border-t p-2">
        <div class="flex items-center gap-2">
          <Avatar class="size-8 shrink-0">
            <AvatarImage
              v-if="userInfo.avatarUrl"
              :src="userInfo.avatarUrl"
              :alt="displayTitle"
            />
            <AvatarFallback class="text-xs">
              {{ avatarFallback }}
            </AvatarFallback>
          </Avatar>
          <span class="text-sm truncate min-w-0 flex-1 group-data-[state=collapsed]:hidden">
            {{ displayNameLabel }}
          </span>
          <Button
            variant="ghost"
            size="icon"
            :title="isInSettingsRoute ? $t('settings.backToChat') : $t('settings.user')"
            :class="isInSettingsRoute ? 'text-primary' : ''"
            @click="onActionButtonClick"
          >
            <FontAwesomeIcon :icon="isInSettingsRoute ? ['fas', 'arrow-left'] : ['fas', 'gear']" />
          </Button>
        </div>
      </SidebarFooter>
    </Sidebar>
  </aside>
</template>

<script setup lang="ts">
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
  Button,
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
} from '@memoh/ui'
import { computed, type Component } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import ChatListMenu from './lists/chat-list-menu.vue'
import SettingsListMenu from './lists/settings-list-menu.vue'
import { useUserStore } from '@/store/User'

const router = useRouter()
const route = useRoute()

type SidebarListKey = 'chat' | 'settings'

const settingsRouteNames = new Set(['settings', 'settings-user', 'bots', 'bot-detail', 'models', 'mcp'])
const sidebarListRegistry: Record<SidebarListKey, Component> = {
  chat: ChatListMenu,
  settings: SettingsListMenu,
}

const currentListKey = computed<SidebarListKey>(() => (
  settingsRouteNames.has(String(route.name ?? '')) ? 'settings' : 'chat'
))
const currentListComponent = computed(() => sidebarListRegistry[currentListKey.value])
const isInSettingsRoute = computed(() => currentListKey.value === 'settings')

const { userInfo } = useUserStore()
const displayNameLabel = computed(() => userInfo.displayName || userInfo.username || userInfo.id || '-')
const displayTitle = computed(() => userInfo.displayName || userInfo.username || userInfo.id || 'User')
const avatarFallback = computed(() => displayTitle.value.slice(0, 2).toUpperCase() || 'U')

function onLogoClick() {
  if (route.name === 'chat') {
    return
  }
  void router.push({ name: 'chat' }).catch(() => undefined)
}

function onActionButtonClick() {
  if (isInSettingsRoute.value) {
    void openChat()
    return
  }
  void openUserSettings()
}

async function openChat() {
  if (route.name === 'chat') {
    return
  }
  await router.push({ name: 'chat' }).catch(() => undefined)
}

async function openUserSettings() {
  if (route.name === 'settings-user') {
    return
  }
  await router.push({ name: 'settings-user' }).catch(() => undefined)
}
</script>
