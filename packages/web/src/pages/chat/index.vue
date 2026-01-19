<template>
  <section class="h-[calc(100vh-calc(var(--spacing)*20))]  max-w-187 gap-8 w-full *:w-full m-auto flex flex-col">
    <!-- <ScrollArea class="flex-none w-full rounded-md border">
      <div class="p-4">
        <h4 class="mb-4 text-sm leading-none font-medium">
          Tags
        </h4>
        <template
          v-for="tag in 1000"
          :key="tag"
        >
          <div class="text-sm">
            {{ tag }}
          </div>
        </template>
      </div>
    </ScrollArea> -->
    <section class="flex-1 h-0">
      <ScrollArea class="max-h-full h-full w-full rounded-md border p-4">
        <ChatList />
      </ScrollArea>
    </section>
    <section class="flex-none relative">
      <Textarea
        class="pb-16 pt-4"
        placeholder="请输入想要获取的内容"
      />
      <section
        class="absolute bottom-0 h-14 px-2 inset-x-0 flex items-center"
      >
        <Button
          variant="default"
          class="ml-auto"
        >
          发送
          <svg-icon
            type="mdi"
            :path="mdiSendOutline"
          />
        </Button>
      </section>
    </section>
  </section>
</template>

<script setup lang="ts">
import {
  ScrollArea, 
  Textarea,
  Button
} from '@memoh/ui'
import SvgIcon from '@jamescoyle/vue-icon'
import { mdiSendOutline } from '@mdi/js'
import ChatList from '@/components/ChatList/index.vue'



</script>