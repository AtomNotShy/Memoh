<template>
  <div class="flex gap-4 items-start">
    <Avatar class="size-8 shrink-0">
      <AvatarImage
        v-if="currentBot?.avatar_url"
        :src="currentBot.avatar_url"
        :alt="assistantLabel"
      />
      <AvatarFallback class="text-xs">
        {{ assistantFallback }}
      </AvatarFallback>
    </Avatar>
    <section class="w-[90%]">
      <sup class="font-semibold">
        {{ assistantLabel }}
      </sup>
      <p class="leading-7 text-muted-foreground break-all">
        <LoadingDots v-if="robotSay.state === 'thinking'" />
        <MarkdownRender
          v-else
          :content="robotSay.description"
          custom-id="chat-answer"
        />
      </p>
    </section>
  </div>
</template>

<script setup lang="ts">
import type { robot } from '@memoh/shared'
import MarkdownRender, { enableKatex, enableMermaid } from 'markstream-vue'
import { Avatar, AvatarFallback, AvatarImage } from '@memoh/ui'
import LoadingDots from '@/components/loading-dots/index.vue'
import { computed } from 'vue'
import { useChatList } from '@/store/chat-list'
import { storeToRefs } from 'pinia'

enableKatex()
enableMermaid()

const props = defineProps<{
  robotSay: robot
}>()

const chatStore = useChatList()
const { botId, bots } = storeToRefs(chatStore)

const currentBot = computed(() => bots.value.find((item) => item.id === botId.value) ?? null)
const assistantLabel = computed(() => (
  currentBot.value?.display_name
  || currentBot.value?.id
  || props.robotSay.type
  || 'Assistant'
))
const assistantFallback = computed(() => assistantLabel.value.slice(0, 2).toUpperCase() || 'AI')
</script>
