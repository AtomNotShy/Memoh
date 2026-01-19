<template>
  <div class="flex flex-col gap-4">
    <template
      v-for="chatItem in chatList"
      :key="chatItem.id"
    >
      <UserChat
        v-if="chatItem.action === 'user'"
        :user-say="chatItem"
      />
      <RobotChat
        v-if="chatItem.action === 'robot'"
        :robot-say="chatItem as robot"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import UserChat from './UserChat/index.vue'
import RobotChat from './RobotChat/index.vue'
import { reactive } from 'vue'
import type { user, robot } from '@memoh/shared'

// 模拟一下数据
const chatList = reactive<(((user | robot) & { action: 'robot' | 'user' }))[]>([{
  description: 'fjiwofwofjewifwe', time: new Date, id: 2, action: 'user'
}, {
  description: 'fjiwofwofjefwfewfwifwe', time: new Date, id: 1000, action: 'robot', type: 'Openai Gpt5'
}, {
  description: 'fjiwofwofjewifwe', time: new Date, id: 2, action: 'user'
}, {
  description: 'fjiwofwofjefwfewfwifwe', time: new Date, id: 1000, action: 'robot', type: 'Openai Gpt5'
}])

</script>