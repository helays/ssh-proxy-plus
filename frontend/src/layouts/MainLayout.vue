<script setup lang="ts">
import {RouterView, useRouter} from 'vue-router'

import {CatchMessageNotification} from '@/components/RespTools.vue'
import {sendRequest} from "@/components/ts/axios";

const router = useRouter()
const currentRoute = window.location.hash.substring(1) || '/';
// const currentRoute = window.location.pathname === '/' ? window.location.pathname : window.location.pathname.substring(1)
const enablePass = (window as any)["enable_pass"];
// 使用计算属性以确保当路由变化时，菜单能响应更新

let computedRoutes:Record<any, any>[]=[];

if(localStorage.getItem("menu")){
  try {
    computedRoutes = JSON.parse(localStorage.getItem("menu") as string);
    // 遍历computedRoutes,如果里面的path==='' 就改成path='/'
    computedRoutes.forEach((item:any) => {
      if (item.path === '') {
        item.path = '/'
      }
    })
  }catch (e){
    CatchMessageNotification(e)
  }
}


const logout = (event:any) => {
  sendRequest({
    request: {
      url: '/api/v1/run.logout',
      method: 'get',
    },
    success: (res:any) => {
      localStorage.removeItem("login_status"); // 退出登录就删除这个key
      router.push("/page/login")
    }
  })
  if (event instanceof PointerEvent) {
    event.stopPropagation();
  }

}

const stopDefaultClickEvent=(event:any)=>{
  // 判断 event是否是 PointerEvent
  if (event instanceof PointerEvent) {
    event.stopPropagation();
  }
}
</script>
<template>
  <div class="common-layout">
    <el-container>
      <el-header class="top-header">
        <el-menu
            :default-active="currentRoute"
            class="el-menu-demo"
            mode="horizontal"
            :ellipsis="false"
            :router="true"
            @select="">
          <el-menu-item index="logo">
            <img src="@/assets/log-108.png" style="width: 58px;" :onclick="stopDefaultClickEvent" alt="隧道代理服务LOGO"/>
          </el-menu-item>

          <el-menu-item v-for="(item,index) in computedRoutes" :index="item.path" :route="item.path">
            {{ item.name }}
          </el-menu-item>

          <div class="flex-grow"/>
          <el-menu-item class="" index="logout" v-if="enablePass=='on'">
            <el-text type="info" @click="logout">退出系统</el-text>
          </el-menu-item>
        </el-menu>
      </el-header>
      <el-main>
        <RouterView/>
      </el-main>
    </el-container>
  </div>
</template>

<style>
.flex-grow {
  flex-grow: 1;
}

.top-header {
  border-bottom: 1px solid var(--el-menu-border-color);
  padding-left: 0;
  padding-right: 0;
}

.el-menu-demo {
  border-bottom: none;
}
</style>


