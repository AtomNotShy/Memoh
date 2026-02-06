<template>
  <section>
    <section class="max-w-187 m-auto">
      <form
        
        @submit="changeSetting"
      >
        <section>
          <section class="flex flex-col">
            <h6 class=" mt-2 mb-2 align-middle">
              <svg-icon
                type="mdi"
                :path="mdiRobotOutline"
                class="inline mr-2 align-[-5px]"
              />模型设置
            </h6>

            <Separator />
          </section>

          <section class="flex flex-col [&_:has(label)]:py-2">
            <FormField
              v-slot="{ componentField }"
              name="chat_model_id"
            >
              <FormItem>
                <Label class="mb-2">
                  Chat Model ID
                </Label>
                <FormControl>
                  <Input v-bind="componentField" />
                </FormControl>
              </FormItem>
            </FormField>
           
            <FormField
              v-slot="{ componentField }"
              name="embedding_model_id"
            >
              <FormItem>
                <Label class="mb-2">
                  Embedding Model ID
                </Label>
                <FormControl>
                  <Input v-bind="componentField" />
                </FormControl>
              </FormItem>
            </FormField>
           
            <FormField
              v-slot="{ componentField }"
              name="memory_model_id"
            >
              <FormItem>
                <Label class="mb-2">
                  Memory Model ID
                </Label>
                <FormControl>
                  <Input v-bind="componentField" />
                </FormControl>
              </FormItem>
            </FormField>
            
            <FormField
              v-slot="{ componentField }"
              name="max_context_load_time"
            >
              <FormItem class="**:[input]:max-w-40!">
                <Label class="mb-2">             
                  Timeout
                </Label>
                <FormControl>
                  <Input v-bind="componentField" />
                </FormControl>
              </FormItem>
            </FormField>
          </section>  
        </section>
        <section class="mt-4">
          <section class="flex flex-col">
            <h6 class=" mt-2 mb-2 align-middle">
              <svg-icon
                type="mdi"
                :path="mdiCog"
                class="inline mr-2 align-[-5px]"
              />显示设置
            </h6>

            <Separator />
          </section>
          <section class="flex flex-col [&_:has(label)]:py-2">
            <FormField
              v-slot="{ componentField }"
              name="language"
            >
              <FormItem class="**:[button]:min-w-40! flex justify-between items-ceneter">
                <Label class="mb-2">
                  语言
                </Label>
                <FormControl>
                  <Select v-bind="componentField">
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        <SelectItem
                          value="zh"
                          @click="$i18n.locale = 'zh'"
                        >
                          中文
                        </SelectItem>
                        <SelectItem
                          value="en"
                          @click="$i18n.locale = 'en'"
                        >
                          English
                        </SelectItem>
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </FormControl>
              </FormItem>
            </FormField>
            <Separator />
            <section class="flex justify-between items-baseline">
              <Label class="mb-4">
                主题
              </Label>

              <Switch
                :model-value="curDark()"
                @update:model-value="() => {
                  toggleMode()
                }
                "
              />
            </section>
          </section>
        </section>
        <section class="mt-4 flex gap-3">
          <Popover>
            <template #default="{ close }">
              <PopoverTrigger as-child>
                <Button
                  class="ml-auto"
                  variant="outline"
                >
                  {{ $t("login.exit") }}
                </Button>
              </PopoverTrigger>
              <PopoverContent class="w-80">
                <div class="grid gap-4">
                  <p class="leading-7 not-first:mt-6  ">
                    确认退出登录?
                  </p>
                  <section class="flex gap-4">
                    <Button
                      variant="outline"
                      class="ml-auto"
                      @click="() => { close() }"
                    >
                      取消
                    </Button>
                    <Button @click="() => { exit(); close() }">
                      确定
                    </Button>
                  </section>
                </div>
              </PopoverContent>
            </template>
          </Popover>
      
          <Button
         
            :disabled="settingLoading"
            type="submit"
          >
            <Spinner v-if="settingLoading" />
            更改
          </Button>
        </section>
      </form>
    </section>
  </section>
</template>

<script setup lang="ts">
import { useMutation, useQuery, useQueryCache } from '@pinia/colada'
import request from '@/utils/request'
import { watch } from 'vue'
import { toTypedSchema } from '@vee-validate/zod'
import z from 'zod'
import { useForm } from 'vee-validate'
import {
  Input,
  FormField,
  FormItem,  
  FormControl,
  Button,
  Select,
  SelectTrigger,
  SelectContent,
  SelectValue,
  SelectGroup,
  SelectItem,
  Label,
  Separator,
  Switch,
  Spinner,
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@memoh/ui'
import SvgIcon from '@jamescoyle/vue-icon'
import { mdiRobotOutline, mdiCog } from '@mdi/js'
import { toast } from 'vue-sonner'
import i18n from '@/i18n'
import { useColorMode } from '@vueuse/core'
import { useRouter } from 'vue-router'
import { useUserStore } from '../../store/User'
const router=useRouter()

const { exitLogin } = useUserStore()
const exit = () => {
  exitLogin()
  router.replace({ name: 'Login' })
}

const mode = useColorMode()
const modeToggleMap: Record<'dark' | 'light', 'dark' | 'light'> = {
  dark: 'light',
  light: 'dark'
}
const toggleMode = () => {
  if (mode.value !== 'auto') {
    mode.value = modeToggleMap[mode.value]
  }
}

const curDark = () => {
  return mode.value==='dark'?true:false
}


const { data: settingData } = useQuery({
  key: ['Setting'],
  query: async () => request({
    url: '/settings',
    method: 'get'
  }).then(fetchSetting => {
    // if(f)
  
    if (fetchSetting?.data?.language&&!i18n.global.availableLocales.includes(fetchSetting?.data?.language)) {
  
      fetchSetting.data.language='zh'
    }
  
    return fetchSetting?.data
  })
})

const formSchema = toTypedSchema(z.object({
  chat_model_id: z.coerce.string(),
  embedding_model_id: z.coerce.string(),
  language: z.coerce.string(),
  max_context_load_time: z.coerce.number().min(1000),
  memory_model_id: z.coerce.string()
}))

const form = useForm({
  validationSchema: formSchema
})

watch(settingData, () => { 
  form.setValues(settingData.value)
  form.values=settingData.value
})

const cacheQuery = useQueryCache()
const { mutate: fetchSetting,isLoading:settingLoading,status } = useMutation({
  mutation: (data:typeof form.values) => request({
    url: '/settings',
    data,
    method:'POST'
  }),
  onSettled: () => {
    cacheQuery.invalidateQueries({
      key: ['Setting']
    })
  }
})

watch(status, () => {
  if (status.value === 'error') {
    toast.error('保存失败') 
  }
})
const changeSetting = form.handleSubmit(async (value) => {

  try { 
    await fetchSetting(value)
  } catch {
    return
  }
})

</script>