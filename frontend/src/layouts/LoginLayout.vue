<script setup lang="ts">
import {reactive, ref} from "vue";
import {ElLoading,ElNotification} from 'element-plus'
import {Lock, View} from '@element-plus/icons-vue'
import {ThenErrorMsgNotification} from '@/components/RespTools.vue'
import { JSEncrypt } from 'jsencrypt'
import {useRouter} from "vue-router";
import {fetchRoutes} from "@/router/routeUtils";
import {sendRequest} from "@/components/ts/axios";

const loginForm = reactive({
  pass: "",
  captcha: "",
})

const timestamp = ref((new Date()).getTime())
const loginDialogVisible = ref(true);
const router = useRouter();
const login = () => {
  if (loginForm.pass === '') {
    ElNotification({
      title: "错误提示",
      message: "密码不能为空",
      type: "error"
    })
    return
  }
  if (loginForm.captcha === '') {
    ElNotification({
      title: "错误提示",
      message: "验证码不能为空",
      type: "error"
    })
    return
  }
  let loading = ElLoading.service({
    lock: true,
    text: '登录中...',
    background: 'rgba(0, 0, 0, 0)',
    fullscreen: false,
  })

  sendRequest({
    request: {
      url: '/api/v1/run.login',
      method: 'get',
    },
    success: (data:any) => {
      if (data.code!==0){
        ThenErrorMsgNotification(data)
        return
      }
      let encrypt = new JSEncrypt();
      encrypt.setPublicKey(data.data);
      sendRequest({
        request: {
          url: '/api/v1/run.login',
          method: 'post',
          data: {
            pass: encrypt.encrypt(loginForm.pass),
            captcha: loginForm.captcha
          }
        },
        success: async (data: any) => {
          if (data.code !== 0) {
            ThenErrorMsgNotification(data)
            return
          }
          ElNotification({
            title: "登录成功",
            message: "正在跳转至控制台...",
            type: "success",
            duration: 1000
          });
          localStorage.setItem("login_status", "login");
          await fetchRoutes(router)
          await router.push('/')
        },
        finally: ()=>{
          loading.close()
        }
      })
    }
  })
}
</script>

<template>
  <div class="login-bg position-absolute"></div>

  <el-dialog v-model="loginDialogVisible" width="400px" :z-index="1" :modal="false" :show-close="false"
             :close-on-click-modal="false" :close-on-press-escape="false" align-center center>
    <template #header>
      <div class="login-header text-center">
        <h1 class="h4">SSH隧道代理服务</h1>
      </div>
    </template>
    <div class="login-form p-4 mt-4" id="login-form">
      <el-form :model="loginForm" label-width="80px" size="default" style="width: 300px;"
               @keydown.enter.prevent="login">
        <el-form-item label="通行证" prop="pass">
          <el-input type="password" v-model="loginForm.pass" autocomplete="off" :prefix-icon="Lock" show-password/>
        </el-form-item>
        <el-form-item label="验证码" prop="captcha">
          <el-row>
            <el-col :span="11">
              <el-input type="text" v-model="loginForm.captcha" autocomplete="off" :prefix-icon="View"/>
            </el-col>
            <el-col :span="13" style="padding-left: 5px;">
              <div style="width: 100%">
                <img style="display: block;width: 106px;height: 40px;cursor: pointer;"
                     :src="'/api/v1/run.captcha?reload='+timestamp" alt="验证码"
                     @click="timestamp=(new Date()).getTime();loginForm.captcha='';">
              </div>
            </el-col>
          </el-row>
        </el-form-item>

      </el-form>
    </div>
    <template #footer>
      <el-button native-type="submit" type="primary" size="default" @click="login">登录</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.login-bg {
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-image: url(@/assets/login-bg.svg);
  background-repeat: no-repeat;
  background-position: center 110px;
  background-size: 100%;
  z-index: -1;
  background-color: #f0f2f5;
}

.login-header h1 {
  margin-bottom: 0;
  font-size: 1.8rem;
  line-height: 40px;
  background: url(@/assets/log-108.png) 0 center no-repeat;
  background-size: contain;
  width: 300px;
  margin-left: auto;
  margin-right: auto;
  font-weight: 300 !important;
}
</style>