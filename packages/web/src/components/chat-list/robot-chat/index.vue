<template>
  <div class="flex gap-4 items-start">
    <div class=" p-2 rounded-full bg-[#F9F9F9] dark:bg-[#666] ">
      <svg-icon
        type="mdi"
        :path="mdiRobotOutline"
      />
    </div>
    <section class="w-[90%]">
      <sup class="font-semibold">
        {{ robotSay.type }}
      </sup>
      <p class="leading-7 text-muted-foreground break-all">
        <template v-if="robotSay.state==='thinking'">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="24"
            height="24"
            viewBox="0 0 24 24"
          >
            <circle
              cx="4"
              cy="12"
              r="3"
              fill="currentColor"
            >
              <animate
                id="SVG7x14Dcom"
                fill="freeze"
                attributeName="opacity"
                begin="0;SVGqSjG0dUp.end-0.25s"
                dur="0.75s"
                values="1;0.2"
              />
            </circle>
            <circle
              cx="12"
              cy="12"
              r="3"
              fill="currentColor"
              opacity="0.4"
            >
              <animate
                fill="freeze"
                attributeName="opacity"
                begin="SVG7x14Dcom.begin+0.15s"
                dur="0.75s"
                values="1;0.2"
              />
            </circle>
            <circle
              cx="20"
              cy="12"
              r="3"
              fill="currentColor"
              opacity="0.3"
            >
              <animate
                id="SVGqSjG0dUp"
                fill="freeze"
                attributeName="opacity"
                begin="SVG7x14Dcom.begin+0.3s"
                dur="0.75s"
                values="1;0.2"
              />
            </circle>
          </svg>
        </template>
        <template v-else>
          <MarkdownRender
            :content="robotSay.description"
            custom-id="chat-answer"
          />
        </template>
      </p>
    </section>
  </div>
</template>

<script setup lang="ts">
import SvgIcon from '@jamescoyle/vue-icon'
import { mdiRobotOutline } from '@mdi/js'
import type { robot } from '@memoh/shared'
import MarkdownRender,{enableKatex,enableMermaid} from 'markstream-vue'
enableKatex()
enableMermaid()
const {robotSay}=defineProps<{
  robotSay: robot
}>()

</script>