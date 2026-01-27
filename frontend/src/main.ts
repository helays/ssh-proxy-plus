import './assets/main.css'

import { createApp } from 'vue'
import ElementPlus from "element-plus"
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'
import 'element-plus/dist/index.css'
import "element-plus/theme-chalk/src/message.scss";

import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(ElementPlus,{
    size: 'small',
    zIndex: 3000,
    locale: zhCn, // 语言包
})

app.use(createPinia())
app.use(router)

app.mount('#app')
