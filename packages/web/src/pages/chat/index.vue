<template>
  <section class="h-[calc(100vh-calc(var(--spacing)*20))] max-w-187 gap-8 w-full *:w-full m-auto flex flex-col">
    <section
      v-if="botId"
      class="flex-none"
    >
      <div class="mx-auto w-full max-w-3xl border rounded-lg px-3 py-2 flex items-center gap-3">
        <Avatar class="size-8 shrink-0">
          <AvatarImage
            v-if="currentBot?.avatar_url"
            :src="currentBot.avatar_url"
            :alt="currentBotTitle"
          />
          <AvatarFallback class="text-xs">
            {{ currentBotFallback }}
          </AvatarFallback>
        </Avatar>
        <div class="min-w-0 flex-1">
          <p class="text-xs text-muted-foreground">
            {{ $t('chat.currentBot') }}
          </p>
          <p class="text-sm font-medium truncate">
            {{ currentBotTitle }}
          </p>
          <p
            v-if="currentBot?.type"
            class="text-xs text-muted-foreground truncate"
          >
            {{ currentBot.type }}
          </p>
        </div>
        <Badge
          v-if="activeChatReadOnly"
          variant="secondary"
        >
          {{ $t('chat.readonly') }}
        </Badge>
      </div>
    </section>

    <section
      v-if="botId"
      class="flex-1 h-0 [&:has(p)]:block! [&:has(p)+section_.logo-title]:hidden [&:has(p)+section]:mt-0! hidden"
    >
      <ScrollArea class="max-h-full h-full w-full rounded-md p-4 **:focus-visible:ring-0!">
        <ChatList />
      </ScrollArea>
    </section>

    <section
      v-if="botId"
      class="flex-none relative m-auto"
    >
      <section class="mb-20 logo-title">
        <h4
          class="scroll-m-20 text-3xl font-semibold tracking-tight text-center"
          style="font-family: 'Source Han Serif CN', 'Noto Serif SC', 'STSong', 'SimSun', serif;"
        >
          <TextGenerateEffect :words="$t('chat.greeting')" />
        </h4>
      </section>

      <Textarea
        v-model="curInputSay"
        class="pb-16 pt-4"
        :placeholder="activeChatReadOnly ? $t('chat.readonlyPlaceholder') : $t('chat.inputPlaceholder')"
        :disabled="activeChatReadOnly"
        @keydown.enter.prevent="send"
      />

      <section class="absolute bottom-0 h-14 px-2 inset-x-0 flex items-center">
        <Button
          variant="default"
          class="ml-auto"
          :disabled="activeChatReadOnly"
          @click="send"
        >
          <template v-if="!loading">
            {{ $t('chat.startChat') }}
            <FontAwesomeIcon :icon="['fas', 'paper-plane']" />
          </template>
          <LoadingDots v-else />
        </Button>
      </section>
    </section>

    <section
      v-else
      class="h-full flex flex-col items-center justify-center text-muted-foreground gap-4"
    >
      <FontAwesomeIcon
        :icon="['fas', 'robot']"
        class="size-12 opacity-20"
      />
      <p>{{ $t('chat.noBot') }}</p>
      <Button @click="router.push({ name: 'bots' })">
        {{ $t('bots.createBot') }}
      </Button>
    </section>
  </section>
</template>

<script setup lang="ts">
import {
  Avatar,
  AvatarFallback,
  AvatarImage,
  Badge,
  ScrollArea,
  Textarea,
  Button,
} from '@memoh/ui'
import ChatList from '@/components/chat-list/index.vue'
import LoadingDots from '@/components/loading-dots/index.vue'
import { computed, onMounted, provide, ref } from 'vue'
import { useChatList } from '@/store/chat-list'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'

const router = useRouter()
const chatSay = ref('')
const curInputSay = ref('')
const chatStore = useChatList()
const {
  loading,
  loadingChats,
  initializing,
  botId,
  bots,
  activeChatReadOnly,
} = storeToRefs(chatStore)

const currentBot = computed(() => bots.value.find((item) => item.id === botId.value) ?? null)
const currentBotTitle = computed(() => currentBot.value?.display_name || currentBot.value?.id || '-')
const currentBotFallback = computed(() => currentBotTitle.value.slice(0, 2).toUpperCase() || 'BT')

provide('chatSay', chatSay)

onMounted(() => {
  void chatStore.initialize().catch(() => undefined)
})

function send() {
  if (!curInputSay.value.trim()) return
  if (activeChatReadOnly.value) return
  if (!loading.value && !loadingChats.value && !initializing.value) {
    chatSay.value = curInputSay.value
    curInputSay.value = ''
  }
}

</script>
