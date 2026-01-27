import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import topLevelAwait from 'vite-plugin-top-level-await';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueJsx(),
    AutoImport({
      resolvers: [ElementPlusResolver()],
    }),
    Components({
      resolvers: [ElementPlusResolver()],
    }),
    topLevelAwait({
      // 可选配置项，例如你可以自定义导出和导入名称
      promiseExportName: '__tla',
      promiseImportName: (i) => `__tla_${i}`,
    }),
  ],
  server:{
    port: 8081,
    proxy:{
      '/api/v1/data.api': {
        target: 'ws://localhost:19009',
        ws: true, // 必须启用WebSocket代理
        changeOrigin: true, // 允许跨域
      },
      "/api/v1":{
        target: 'http://localhost:19009',
        changeOrigin: true, // 允许跨域
      },
    }
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  build: {
    outDir: '../dist/html', // 将输出目录设置为'build'，根据需要自定义
    emptyOutDir: true, // 强制清空目录
    assetsDir: 'assets', // 资源文件输出目录
    rollupOptions: {
      // ...其他Rollup配置项
      output: {
        manualChunks: {
          'element-plus': ['element-plus', 'vue', 'vue-router'],
        },
      },
    },
  },
  // // esbuild 转换设置，这里可以调整目标环境
  // esbuild: {
  //   target: 'es2022', // 尝试将目标环境改为更高版本，比如 "es2022" 或 "esnext"
  // },
})