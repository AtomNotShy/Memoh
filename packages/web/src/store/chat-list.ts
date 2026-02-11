import { defineStore } from 'pinia'
import { computed, reactive, ref, watch } from 'vue'
import { useLocalStorage } from '@vueuse/core'
import type { user, robot } from '@memoh/shared'
import {
  createChat,
  deleteChat as requestDeleteChat,
  type Bot as ChatBot,
  type ChatSummary,
  fetchBots,
  fetchChatMessages,
  fetchChats,
  extractAssistantTexts,
  extractPersistedMessageText,
  streamChatMessage,
} from '@/composables/api/useChat'

export const useChatList = defineStore('chatList', () => {
  const chatList = reactive<(user | robot)[]>([])
  const chats = ref<ChatSummary[]>([])
  const loading = ref(false)
  const loadingChats = ref(false)
  const initializing = ref(false)
  const botId = useLocalStorage<string | null>('chat-bot-id', null)
  const chatId = useLocalStorage<string | null>('chat-id', null)
  const bots = ref<ChatBot[]>([])
  const participantChats = computed(() =>
    chats.value.filter((item) => (item.access_mode ?? 'participant') === 'participant'),
  )
  const observedChats = computed(() =>
    chats.value.filter((item) => item.access_mode === 'channel_identity_observed'),
  )
  const activeChat = computed(() =>
    chats.value.find((item) => item.id === chatId.value) ?? null,
  )
  const activeChatReadOnly = computed(() => activeChat.value?.access_mode === 'channel_identity_observed')

  // Watch for botId changes to re-initialize
  watch(botId, (newBotId) => {
    if (newBotId) {
      void initialize()
    } else {
      chats.value = []
      chatId.value = null
      replaceMessages([])
    }
  })

  const nextId = () => `${Date.now()}-${Math.floor(Math.random() * 1000)}`

  const resolveBotIdentityLabel = (targetBotID?: string | null) => {
    const activeBotID = targetBotID ?? botId.value
    if (!activeBotID) {
      return 'Assistant'
    }
    const currentBot = bots.value.find((item) => item.id === activeBotID)
    return currentBot?.display_name?.trim() || currentBot?.id || 'Assistant'
  }

  const addUserMessage = (text: string) => {
    chatList.push({
      description: text,
      time: new Date(),
      action: 'user',
      id: nextId(),
    })
  }

  const addRobotMessage = (text: string, state: robot['state'] = 'complete') => {
    const id = nextId()
    chatList.push({
      description: text,
      time: new Date(),
      action: 'robot',
      id,
      type: resolveBotIdentityLabel(),
      state,
    })
    return id
  }

  const updateRobotMessage = (id: string, patch: Partial<robot>) => {
    const target = chatList.find(
      (item): item is robot => item.action === 'robot' && String(item.id) === id,
    )
    if (target) {
      Object.assign(target, patch)
    }
  }

  const ensureBot = async () => {
    try {
      const botsList = await fetchBots()
      bots.value = botsList
      if (!botsList.length) {
        botId.value = null
        return null
      }
      // If we have a persisted botId and it's still valid, use it
      if (botId.value && botsList.some((b) => b.id === botId.value)) {
        return botId.value
      }
      // Otherwise default to the first bot
      botId.value = botsList[0]!.id
      return botId.value
    } catch (error) {
      console.error('Failed to fetch bots:', error)
      return botId.value // Fallback to whatever we have
    }
  }

  const replaceMessages = (items: (user | robot)[]) => {
    chatList.splice(0, chatList.length, ...items)
  }

  const toChatItem = (raw: Awaited<ReturnType<typeof fetchChatMessages>>[number]): user | robot | null => {
    if (raw.role !== 'user' && raw.role !== 'assistant') {
      return null
    }

    const text = extractPersistedMessageText(raw)
    if (!text) {
      return null
    }

    const createdAt = raw.created_at ? new Date(raw.created_at) : new Date()
    const time = Number.isNaN(createdAt.getTime()) ? new Date() : createdAt
    const itemID = raw.id || nextId()

    if (raw.role === 'user') {
      return {
        description: text,
        time,
        action: 'user',
        id: itemID,
      }
    }

    return {
      description: text,
      time,
      action: 'robot',
      id: itemID,
      type: resolveBotIdentityLabel(raw.bot_id || botId.value),
      state: 'complete',
    }
  }

  const loadMessages = async (targetChatID: string) => {
    const rows = await fetchChatMessages(targetChatID)
    const items = rows
      .map(toChatItem)
      .filter((item): item is user | robot => item !== null)
    replaceMessages(items)
  }

  const initialize = async () => {
    if (initializing.value) {
      return
    }

    initializing.value = true
    loadingChats.value = true
    try {
      const currentBotID = await ensureBot()
      if (!currentBotID) {
        chats.value = []
        chatId.value = null
        replaceMessages([])
        return
      }
      const visibleChats = await fetchChats(currentBotID)
      chats.value = visibleChats

      if (visibleChats.length === 0) {
        chatId.value = null
        replaceMessages([])
        return
      }

      const activeChatID = chatId.value && visibleChats.some((item) => item.id === chatId.value)
        ? chatId.value
        : visibleChats[0]!.id
      chatId.value = activeChatID
      await loadMessages(activeChatID)
    } finally {
      loadingChats.value = false
      initializing.value = false
    }
  }

  const selectBot = async (targetBotID: string) => {
    if (botId.value === targetBotID) {
      return
    }
    botId.value = targetBotID
    chatId.value = null
    await initialize()
  }

  const createNewChat = async () => {
    loadingChats.value = true
    try {
      const currentBotID = await ensureBot()
      if (!currentBotID) return
      const created = await createChat(currentBotID)
      chats.value = [created, ...chats.value.filter((item) => item.id !== created.id)]
      chatId.value = created.id
      replaceMessages([])
    } finally {
      loadingChats.value = false
    }
  }

  const removeChat = async (targetChatID: string) => {
    const deletingChatID = targetChatID.trim()
    if (!deletingChatID) {
      return
    }

    loadingChats.value = true
    try {
      await requestDeleteChat(deletingChatID)
      const remainingChats = chats.value.filter((item) => item.id !== deletingChatID)
      chats.value = remainingChats

      if (chatId.value !== deletingChatID) {
        return
      }

      if (remainingChats.length === 0) {
        chatId.value = null
        replaceMessages([])
        return
      }

      const nextChatID = remainingChats[0]!.id
      chatId.value = nextChatID
      await loadMessages(nextChatID)
    } finally {
      loadingChats.value = false
    }
  }

  const selectChat = async (targetChatID: string) => {
    const nextChatID = targetChatID.trim()
    if (!nextChatID || nextChatID === chatId.value) {
      return
    }

    chatId.value = nextChatID
    loadingChats.value = true
    try {
      await loadMessages(nextChatID)
    } finally {
      loadingChats.value = false
    }
  }

  const ensureActiveChat = async () => {
    if (chatId.value) {
      return
    }
    const currentBotID = botId.value ?? await ensureBot()
    if (!currentBotID) {
      throw new Error('Bot not ready')
    }
    const created = await createChat(currentBotID)
    chats.value = [created, ...chats.value.filter((item) => item.id !== created.id)]
    chatId.value = created.id
    replaceMessages([])
  }

  const touchChat = (targetChatID: string) => {
    const index = chats.value.findIndex((item) => item.id === targetChatID)
    if (index < 0) {
      return
    }
    const [target] = chats.value.splice(index, 1)
    if (!target) {
      return
    }
    target.updated_at = new Date().toISOString()
    chats.value.unshift(target)
  }

  const sendMessage = async (text: string) => {
    const trimmed = text.trim()
    if (!trimmed) return

    loading.value = true
    let thinkingId: string | null = null
    try {
      await ensureActiveChat()
      const activeChatID = chatId.value!
      if (activeChatReadOnly.value) {
        throw new Error('Chat is read-only')
      }
      addUserMessage(trimmed)

      thinkingId = addRobotMessage('', 'thinking')
      const currentThinkingID = thinkingId
      let streamedText = ''
      const finalResponse = await streamChatMessage(activeChatID, trimmed, (delta) => {
        if (!delta) {
          return
        }
        streamedText += delta
        updateRobotMessage(currentThinkingID, {
          description: streamedText,
          state: 'generate',
        })
      })

      if (streamedText.trim()) {
        updateRobotMessage(currentThinkingID, {
          description: streamedText.trim(),
          state: 'complete',
        })
        touchChat(activeChatID)
        return
      }

      const assistantTexts = extractAssistantTexts(finalResponse?.messages ?? [])
      if (assistantTexts.length === 0) {
        updateRobotMessage(currentThinkingID, {
          description: 'No textual response.',
          state: 'complete',
        })
        touchChat(activeChatID)
        return
      }

      updateRobotMessage(currentThinkingID, {
        description: assistantTexts[0]!,
        state: 'complete',
      })
      for (const textItem of assistantTexts.slice(1)) {
        addRobotMessage(textItem)
      }
      touchChat(activeChatID)
    } catch (error) {
      const reason = error instanceof Error ? error.message : 'Unknown error'
      if (thinkingId) {
        updateRobotMessage(thinkingId, {
          description: `Failed to send message: ${reason}`,
          state: 'complete',
        })
      } else {
        addRobotMessage(`Failed to send message: ${reason}`)
      }
      throw error
    } finally {
      loading.value = false
    }
  }

  return {
    chatList,
    chats,
    participantChats,
    observedChats,
    chatId,
    botId,
    bots,
    activeChat,
    activeChatReadOnly,
    loading,
    loadingChats,
    initializing,
    initialize,
    selectBot,
    selectChat,
    createNewChat,
    removeChat,
    deleteChat: removeChat,
    sendMessage,
  }
})
