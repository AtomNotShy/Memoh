<template>
  <section class="h-full max-w-7xl mx-auto p-6">
    <section class="min-w-0">
      <template v-if="isGeneralSection">
        <div class="max-w-3xl">
          <h6 class="mb-2 flex items-center">
            <FontAwesomeIcon
              :icon="['fas', 'gear']"
              class="mr-2"
            />
            {{ $t('settings.display') }}
          </h6>
          <Separator />

          <div class="mt-4 space-y-4">
            <div class="flex items-center justify-between">
              <Label>{{ $t('sidebar.bots') }}</Label>
              <Select
                :model-value="botId ?? ''"
                :disabled="loadingChats"
                @update:model-value="(v) => v && onSelectBot(String(v))"
              >
                <SelectTrigger class="w-64">
                  <SelectValue :placeholder="$t('bots.title')" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem
                      v-for="bot in bots"
                      :key="bot.id"
                      :value="bot.id"
                    >
                      {{ bot.display_name || bot.id }}
                    </SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>

            <Separator />

            <div class="flex items-center justify-between">
              <Label>{{ $t('settings.language') }}</Label>
              <Select
                :model-value="language"
                @update:model-value="(v) => v && setLanguage(v as Locale)"
              >
                <SelectTrigger class="w-40">
                  <SelectValue :placeholder="$t('settings.languagePlaceholder')" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem value="zh">
                      {{ $t('settings.langZh') }}
                    </SelectItem>
                    <SelectItem value="en">
                      {{ $t('settings.langEn') }}
                    </SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>

            <Separator />

            <div class="flex items-center justify-between">
              <Label>{{ $t('settings.theme') }}</Label>
              <Select
                :model-value="theme"
                @update:model-value="(v) => v && setTheme(v as 'light' | 'dark')"
              >
                <SelectTrigger class="w-40">
                  <SelectValue :placeholder="$t('settings.themePlaceholder')" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem value="light">
                      {{ $t('settings.themeLight') }}
                    </SelectItem>
                    <SelectItem value="dark">
                      {{ $t('settings.themeDark') }}
                    </SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>
          </div>

        </div>
      </template>

      <section
        v-else
        class="relative min-h-[70vh]"
      >
        <RouterView />
      </section>
    </section>
  </section>
</template>

<script setup lang="ts">
import {
  Select,
  SelectTrigger,
  SelectContent,
  SelectValue,
  SelectGroup,
  SelectItem,
  Label,
  Separator,
} from '@memoh/ui'
import { computed, onMounted } from 'vue'
import { useRoute, RouterView } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useSettingsStore } from '@/store/settings'
import { useChatList } from '@/store/chat-list'
import type { Locale } from '@/i18n'

const route = useRoute()
const settingsStore = useSettingsStore()
const chatStore = useChatList()
const { language, theme } = storeToRefs(settingsStore)
const { setLanguage, setTheme } = settingsStore
const { bots, botId, loadingChats } = storeToRefs(chatStore)

onMounted(() => {
  void chatStore.initialize().catch(() => undefined)
})

const isGeneralSection = computed(() => route.name === 'settings')

function onSelectBot(nextBotID: string) {
  if (!nextBotID) {
    return
  }
  void chatStore.selectBot(nextBotID).catch(() => undefined)
}
</script>
