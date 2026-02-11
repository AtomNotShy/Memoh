<template>
  <SidebarGroup>
    <SidebarGroupLabel :class="collapsedHiddenClass">
      <span>{{ $t('sidebar.settings') }}</span>
    </SidebarGroupLabel>
    <SidebarGroupContent>
      <SidebarMenu>
        <SidebarMenuItem
          v-for="item in items"
          :key="item.key"
        >
          <SidebarMenuButton
            :is-active="item.active"
            :tooltip="item.label"
            @click="openSection(item.routeName)"
          >
            <FontAwesomeIcon :icon="item.icon" />
            <span :class="collapsedHiddenClass">{{ item.label }}</span>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarGroupContent>
  </SidebarGroup>
</template>

<script setup lang="ts">
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@memoh/ui'
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import type { SidebarListProps } from './types'

const props = withDefaults(defineProps<SidebarListProps>(), {
  collapsible: false,
})

const route = useRoute()
const router = useRouter()
const { t } = useI18n()

const collapsedHiddenClass = computed(() => (
  props.collapsible ? 'group-data-[state=collapsed]:hidden' : ''
))

const items = computed(() => [
  {
    key: 'user',
    label: t('settings.user'),
    routeName: 'settings-user',
    icon: ['fas', 'user'] as const,
    active: route.name === 'settings-user',
  },
  {
    key: 'general',
    label: t('settings.display'),
    routeName: 'settings',
    icon: ['fas', 'gear'] as const,
    active: route.name === 'settings',
  },
  {
    key: 'bots',
    label: t('sidebar.bots'),
    routeName: 'bots',
    icon: ['fas', 'robot'] as const,
    active: route.name === 'bots' || route.name === 'bot-detail',
  },
  {
    key: 'models',
    label: t('sidebar.models'),
    routeName: 'models',
    icon: ['fas', 'cubes'] as const,
    active: route.name === 'models',
  },
  {
    key: 'mcp',
    label: t('sidebar.mcp'),
    routeName: 'mcp',
    icon: ['fas', 'plug'] as const,
    active: route.name === 'mcp',
  },
])

function openSection(routeName: string) {
  void router.push({ name: routeName }).catch(() => undefined)
}
</script>

