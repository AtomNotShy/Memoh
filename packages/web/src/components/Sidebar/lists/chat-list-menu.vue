<template>
  <section>
    <div :class="['px-3 pb-2', collapsedHiddenClass]">
      <DropdownMenu>
        <DropdownMenuTrigger as-child>
          <Button
            class="w-full justify-between"
            :disabled="newChatDisabled"
          >
            <span>{{ $t('chat.newChat') }}</span>
            <FontAwesomeIcon
              :icon="['fas', 'chevron-down']"
              class="size-3 opacity-70"
            />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent
          align="start"
          class="w-64"
        >
          <DropdownMenuItem
            v-for="bot in bots"
            :key="bot.id"
            @click="onCreateChat(bot.id)"
          >
            <span class="truncate">{{ bot.display_name || bot.id }}</span>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>

    <SidebarGroup>
      <SidebarGroupLabel :class="collapsedHiddenClass">
        <span>{{ $t('chat.history') }}</span>
      </SidebarGroupLabel>
      <SidebarGroupContent>
        <SidebarMenu v-if="botId">
          <div
            v-if="participantChats.length > 0"
            :class="['px-2 pb-1 text-[11px] text-muted-foreground uppercase tracking-wide', collapsedHiddenClass]"
          >
            {{ $t('chat.historyParticipant') }}
          </div>
          <SidebarMenuItem
            v-for="(chat, index) in participantChats"
            :key="chat.id"
            class="group/chat relative"
          >
            <SidebarMenuButton
              :is-active="chatId === chat.id"
              :tooltip="chatLabel(chat, index)"
              class="pr-8"
              @click="onSelectChat(chat.id)"
            >
              <FontAwesomeIcon :icon="['far', 'comment']" />
              <span class="truncate">{{ chatLabel(chat, index) }}</span>
            </SidebarMenuButton>

            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button
                  variant="ghost"
                  size="icon"
                  :class="[
                    'size-6 absolute right-1 top-1/2 -translate-y-1/2 opacity-0 group-hover/chat:opacity-100 transition-opacity',
                    collapsedHiddenClass,
                  ]"
                  @click.stop
                >
                  <FontAwesomeIcon :icon="['fas', 'ellipsis-vertical']" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem
                  class="text-destructive focus:text-destructive"
                  @click="onDeleteChat(chat.id)"
                >
                  {{ $t('common.delete') }}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>

          <div
            v-if="observedChats.length > 0"
            :class="['px-2 pt-3 pb-1 text-[11px] text-muted-foreground uppercase tracking-wide', collapsedHiddenClass]"
          >
            {{ $t('chat.historyObserved') }}
          </div>
          <SidebarMenuItem
            v-for="(chat, index) in observedChats"
            :key="chat.id"
            class="group/chat relative"
          >
            <SidebarMenuButton
              :is-active="chatId === chat.id"
              :tooltip="chatLabel(chat, index)"
              class="pr-2"
              @click="onSelectChat(chat.id)"
            >
              <FontAwesomeIcon :icon="['far', 'comment']" />
              <span class="truncate">{{ chatLabel(chat, index) }}</span>
              <span class="ml-auto text-[10px] text-muted-foreground">{{ $t('chat.readonly') }}</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  </section>
</template>

<script setup lang="ts">
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  Button,
} from '@memoh/ui'
import { computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useChatList } from '@/store/chat-list'
import { storeToRefs } from 'pinia'
import type { ChatSummary } from '@/composables/api/useChat'
import type { SidebarListProps } from './types'

const props = withDefaults(defineProps<SidebarListProps>(), {
  collapsible: false,
})

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const chatStore = useChatList()
const {
  botId,
  bots,
  participantChats,
  observedChats,
  chatId,
  loadingChats,
  initializing,
} = storeToRefs(chatStore)

const collapsedHiddenClass = computed(() => (
  props.collapsible ? 'group-data-[state=collapsed]:hidden' : ''
))
const newChatDisabled = computed(() => (
  loadingChats.value || initializing.value || bots.value.length === 0
))

onMounted(() => {
  void chatStore.initialize().catch(() => undefined)
})

function chatLabel(chat: ChatSummary, index: number) {
  const title = chat.title?.trim()
  if (title) {
    return title
  }
  return `${t('sidebar.chat')} ${index + 1}`
}

async function onCreateChat(targetBotID: string) {
  if (!targetBotID) {
    return
  }
  try {
    await chatStore.selectBot(targetBotID)
    await chatStore.createNewChat()
    if (route.name !== 'chat') {
      await router.push({ name: 'chat' })
    }
  } catch {
    return
  }
}

async function onSelectChat(targetChatID: string) {
  if (!targetChatID) {
    return
  }
  try {
    await chatStore.selectChat(targetChatID)
    if (route.name !== 'chat') {
      await router.push({ name: 'chat' })
    }
  } catch {
    return
  }
}

async function onDeleteChat(targetChatID: string) {
  if (!targetChatID) {
    return
  }
  const removeAction = chatStore.removeChat
  if (typeof removeAction !== 'function') {
    return
  }
  try {
    await removeAction(targetChatID)
  } catch {
    return
  }
}
</script>

