<template>
  <aside class="[&_[data-state=collapsed]_.title-container]:hidden">
    <Sidebar
      collapsible="icon"
    >
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>        
            <img
              src="../../../public/logo.png"
              width="80"
              class="m-auto"
              alt="logo.png"
            >
            <h4 class="scroll-m-20 text-xl font-semibold tracking-tight text-center text-muted-foreground title-container">
              Memoh
            </h4>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>
            对话操作
          </SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <Collapsible
                v-for="sidebarItem in sidebarInfo"
                :key="sidebarItem.title"
                class="group/collapsible"
              >
                <SidebarMenuItem>
                  <CollapsibleTrigger as-child>
                    <SidebarMenuButton :tooltip="sidebarItem.title">
                      <svg-icon
                        type="mdi"
                        :path="sidebarItem.icon"
                      />
                      <span>{{ sidebarItem.title }}</span>
                    </SidebarMenuButton>
                  </CollapsibleTrigger>
                </SidebarMenuItem>
              </Collapsible>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
     
      <SidebarRail />
    </Sidebar>
  </aside>
</template>
<script setup lang="ts">
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,  
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,  
  SidebarRail, 
  CollapsibleTrigger,
  Collapsible
  
} from '@memoh/ui'
import { reactive, inject } from 'vue'
import SvgIcon from '@jamescoyle/vue-icon'
import { mdiRobot, mdiChatOutline, mdiCogBox, mdiListBox, mdiHome } from '@mdi/js'


const open=inject('sideBarIsOpen')

const sidebarInfo = reactive([{
  title: '创建对话',
  path: '/',
  icon: mdiChatOutline
}, {
  title: '主页',
  path: '/',
  icon: mdiHome
},
  {
  title: '模型配置',
  path: '/',
  icon: mdiRobot
  }, {
  title: '环境设置',
  path: '/',
  icon: mdiCogBox
  }, {
  title: '平台',
  path: '/',
  icon: mdiListBox
}])
</script>