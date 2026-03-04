<script setup lang="ts">
import {reactive, ref} from 'vue';
import {deepCopyJson, handleCopy} from "@/components/ts/utils";
import TableTool from "@/components/page/tableTool.vue";
import {sendRequest} from "@/components/ts/axios";
import { Plus, Refresh} from "@element-plus/icons-vue";
import {ElMessage} from "element-plus";


const defaultSearchForm={
  address:'',
  type:'',
  is_alive: '',
  port_open: '',
}
const searchForm = reactive(deepCopyJson(defaultSearchForm));
const refresh = ref(Date.now());
// 格式化延迟（time.Duration 是纳秒数）
const formatLatency=(row: any, column: any, cellValue: any)=> {
  if (!cellValue) return '-'

  const ns = Number(cellValue) // time.Duration 是纳秒数

  if (ns < 1000) {
    return ns + ' ns'
  } else if (ns < 1000 * 1000) {
    return (ns / 1000).toFixed(2) + ' μs'
  } else if (ns < 1000 * 1000 * 1000) {
    return (ns / (1000 * 1000)).toFixed(2) + ' ms'
  } else {
    return (ns / (1000 * 1000 * 1000)).toFixed(2) + ' s'
  }
}

// 格式化速率（字节/毫秒）
const formatSpeed=(row: any, column: any, cellValue: number)=> {
  if (!cellValue && cellValue !== 0) return '-'

  const bytesPerMs = Number(cellValue)
  const bytesPerSecond = bytesPerMs * 1000

  if (bytesPerSecond < 1024) {
    return bytesPerSecond.toFixed(2) + ' B/s'
  } else if (bytesPerSecond < 1024 * 1024) {
    return (bytesPerSecond / 1024).toFixed(2) + ' KB/s'
  } else if (bytesPerSecond < 1024 * 1024 * 1024) {
    return (bytesPerSecond / (1024 * 1024)).toFixed(2) + ' MB/s'
  } else {
    return (bytesPerSecond / (1024 * 1024 * 1024)).toFixed(2) + ' GB/s'
  }
}

const formatScore=(row: any, column: any, cellValue: null | undefined)=> {
  if (cellValue === undefined || cellValue === null) return '-'
  return Number(cellValue).toFixed(2)
}

const proxyTest = (id:any) =>{
  sendRequest({
    loading: true,
    request:{
      url: '/api/v1/proxy/test',
      method: 'post',
      params: {id: id}
    },
    success: (res:any) => {
      refresh.value = Date.now()
    }
  })
}

const proxyDelete=(id:any) =>{
  sendRequest({
    request:{
      url: '/api/v1/proxy/delete',
      method: 'post',
      params: {id: id}
    },
    success: (res:any) => {
      refresh.value = Date.now()
    }
  })
}

const updateBest=()=>{
  sendRequest({
    request:{
      url: '/api/v1/proxy/update-best',
      method: 'post',
    }
  })
}

const dialogVisible = ref(false);
interface formData {
  address:string,
}
const formDataDefault:formData = {
  address: '',
}
const formData=reactive(deepCopyJson(formDataDefault));

const proxyCreate=()=>{
  if (!formData.address){
    ElMessage.error('请填写代理地址')
    return
  }
  // 代理地址格式验证
  const proxyRegex = /^(socks5|socks4|https|http):\/\/((?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)|localhost|[a-zA-Z0-9\-\.]+(?:\.[a-zA-Z]{2,})):(\d{1,5})$/

  if (!proxyRegex.test(formData.address)) {
    ElMessage.error('代理地址格式不正确，正确格式为：(socks5|socks4|https|http)://ip:port')
    return
  }

  sendRequest({
    request:{
      url: '/api/v1/proxy/create',
      method: 'post',
      data: {address: formData.address}
    },
    success: (res:any) => {
      refresh.value = Date.now()
      dialogVisible.value = false
    }
  })
}

</script>

<template>

  <div style="margin-bottom: 20px;">
    <el-button type="primary" size="small" :icon="Plus" @click="dialogVisible = true">添加代理</el-button>
    <el-button type="primary" size="small" :icon="Refresh" @click="updateBest">更新最优代理</el-button>
  </div>

  <el-dialog v-model="dialogVisible" title="添加代理" width="500px" :append-to-body="true" :close-on-click-modal="false"
  :draggable="true" :destroy-on-close="true">
    <el-form label-width="auto" :model="formData" onsubmit="return false;">
      <el-form-item label="代理地址" required>
        <el-input v-model="formData.address" clearable placeholder="代理地址" autocomplete="on" style="width: 300px"/>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" @click="proxyCreate">保存</el-button>
    </template>

  </el-dialog>


  <table-tool
      list-api="/api/v1/proxy/list"
      :enable-serial="false"
      :search-model="searchForm"
      :refresh="refresh"
      :enable-search="true"
      search-size="small"
      page-field="pageNo"
      ref="tableRef"
      page-size-field="pageSize">
    <template #search>
      <el-form-item label="代理地址">
        <el-input v-model="searchForm.address" clearable placeholder="代理地址" autocomplete="on" style="width: 130px"/>
      </el-form-item>
      <el-form-item label="代理类型">
        <el-select v-model="searchForm.type" clearable placeholder="代理类型" style="min-width: 88px">
          <el-option label="socks5" value="socks5"/>
          <el-option label="socks4" value="socks4"/>
          <el-option label="https" value="https"/>
          <el-option label="http" value="http"/>
        </el-select>
      </el-form-item>

      <el-form-item label="状态">
        <el-select v-model="searchForm.is_alive" clearable placeholder="状态" style="min-width: 88px">
          <el-option label="正常" value="1"/>
          <el-option label="不可用" value="0"/>
        </el-select>
      </el-form-item>
      <el-form-item label="端口状态">
        <el-select v-model="searchForm.port_open" clearable placeholder="状态" style="min-width: 88px">
          <el-option label="正常" value="1"/>
          <el-option label="不可用" value="0"/>
        </el-select>
      </el-form-item>
    </template>

    <template #default>
      <el-table-column label="代理地址">
        <template v-slot="{ row }">
          <el-tag type="primary" @dblclick="handleCopy(row.address)" class="unselect cur-pointer">{{ row.address }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="latency" label="延迟" :formatter="formatLatency"/>
      <el-table-column prop="speed" label="速率" :formatter="formatSpeed"/>
      <el-table-column prop="score" label="综合评分" :formatter="formatScore"/>
      <el-table-column prop="last_check" label="最后检测时间"/>
      <el-table-column prop="is_alive" label="是否可用">
        <template v-slot="{ row }">
          <el-tag v-if="row.is_alive" type="success">正常</el-tag>
          <el-tag v-else type="danger">不可用</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="port_open" label="端口状态">
        <template v-slot="{ row }">
          <el-tag v-if="row.port_open" type="success">正常</el-tag>
          <el-tag v-else type="danger">不可用</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="message" label="备注"/>

      <el-table-column label="操作">
        <template v-slot="{ row }">
          <el-popconfirm title="确认开始测试？" confirm-button-text="确认" cancel-button-text="取消" @confirm="proxyTest(row.id)">
            <template #reference>
              <el-button type="primary" text size="small">测试</el-button>
            </template>
          </el-popconfirm>

          <el-popconfirm title="确认删除？" confirm-button-text="确认" cancel-button-text="取消" @confirm="proxyDelete(row.id)">
            <template #reference>
              <el-button type="danger" text size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </template>
  </table-tool>

</template>

<style scoped>

</style>