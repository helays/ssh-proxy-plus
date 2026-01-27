<script setup lang="ts">
import {reactive, onBeforeMount, h, onBeforeUnmount,ref} from "vue";
import {Plus} from '@element-plus/icons-vue'
import {ElNotification,ElLoading} from 'element-plus'

import qs from 'qs';
import {sendRequest} from "@/components/ts/axios";
import {handleCopy} from "@/components/ts/utils";

interface connectStatus {
  text:string,
  type: string,
}
const logs = reactive<string[]>([])
const ListData = reactive<Record<any, any>[]>([]);
const dialogVisibleLogs = ref(false);
const dialogVisible = reactive({
  form: false,
  title: "",
  loadding: true,
  remote: true,
})
const formModel = reactive<any>({
  connect: "",
  id: "",
  listen: "",
  lname: "",
  passwd: "",
  pid: "",
  remote: "",
  saddr: "",
  type: "",
  user: "",
})
const connectTypeMap: Record<string, string> = {
  "L": "本地代理",
  "R": "远程代理",
  "D": "动态代理",
  "H":"HTTP代理"
}
const connectStatusMap: Record<number, connectStatus> = {
  0: {text: "连接失败", type: "danger"},
  1: {text: "新增", type: "info"},
  2: {text: "连接中", type: "primary"},
  3: {text: "连接成功", type: "success"},
  4: {text: "中断连接", type: "warning"},
}
let ws:WebSocket;
let setTimeoutId: number, setIntervalId: any;

const socket = () => {
  ws = new WebSocket("ws://" + window.location.host + "/api/v1/data.api");
  ws.onopen = function (evt) {
    clearTimeout(setTimeoutId);
    setIntervalId = setInterval(function () {
      ws.send('1');
    }, 4000);
  };
  ws.onmessage = function (evt: any) {
    let resp = JSON.parse(evt.data);
    switch (resp['action']) {
      case 'list':
        dialogVisible.loadding = false;
        ListData.splice(0, ListData.length);
        listDataHandler(resp['data'], 1);
        break;
      case 'status':
        listDataStatusHandler(ListData, resp['data']);
        break;
      case "log":
        logs.push(resp['data']);
        break;
    }
  }.bind(this);
  ws.onclose = function () {
    clearInterval(setIntervalId);
    setTimeoutId = setTimeout(function () {
      socket();
    },2000);
  }
}


const listDataHandler = (data: Record<any, any>[], index: number) => {
  if (!data){
    return [];
  }
  data.forEach(function (item: Record<any, any>) {
    item['connectType'] = (connectTypeMap[item["connect"]]) as string;
    item["remoteText"] = item["remote"];
    if (item["connect"] === "D") {
      item["remoteText"] = "动态代理"
    }else if (item["connect"] === "H"){
      item["remoteText"] = "HTTP代理"
    }
    if (item["son"] && item["son"].length > 0) {
      item["son"] = listDataHandler(item["son"], index + 1);
    }
    if (index === 1) {
      ListData.push(item);
    }
  })
  if (index === 1) {
    return [];
  }
  return data;
}

const listDataStatusHandler = (data: Record<any, any>[], statusMap: Record<string, number>) => {
  data.forEach(function (item:Record<any, any>) {
    let status = statusMap[item['id']];
    item["connectStatus"] = connectStatusMap[status];
    if (!item["connectStatus"]){
      item["connectStatus"]=(item.active==='Y' ? {text: "'无状态'", type: "danger"} : connectStatusMap[4]);
    }
    if (item["son"] && item["son"].length > 0) {
      listDataStatusHandler(item['son'], statusMap);
    }
  });
}

const openForm = (inputData: any) => {
  dialogVisible.title = "添加隧道"
  if (inputData["id"]) {
    dialogVisible.title = "编辑隧道"
  }
  dialogVisible.remote = !(inputData["connect"] === 'D' || inputData["connect"] === 'H');
  dialogVisible.form = true
  for (let key in formModel) {
    formModel[key] = inputData[key] ?? '';
  }
  formModel.passwd = '';
}

const connectChange = (val: any) => {
  dialogVisible.remote = (val !== 'D' && val !== 'H');
}

const deleteData = (id: any) => {
  const loading =ElLoading.service({background: 'rgba(0, 0, 0, 0)', fullscreen: false,lock:true})
  sendRequest({
    request: {
      url: "/api/v1/edit.api",
      method: 'DELETE',
      params: {
        id: id,
      }
    },
    disableSuccessNotice: true,
    success: function (resp:any) {
      if (resp.code === 0) {
        ElNotification({
          title: '成功',
          type: 'success',
          message: h('i', {style: 'color: #67C23A'}, '删除成功'),
        })
      } else {
        ElNotification({
          title: '异常',
          type: 'error',
          message: h('i', {style: 'color: #F56C6C'}, resp.msg),
        })
      }
    },
    finally: function () {
      loading.close();
      dialogVisible.form = false;
    }

  })
}

const changeConnect = (id:string) => {
  const loading =ElLoading.service({background: 'rgba(0, 0, 0, 0)', fullscreen: false,lock:true})
  sendRequest({
    request: {
      url: "/api/v1/edit.api",
      method: 'PATCH',
      params: {id: id}
    },
    finally:()=>{
      loading.close();
    }
  })
}

const submit = () => {
  const loading =ElLoading.service({background: 'rgba(0, 0, 0, 0)', fullscreen: true,lock:true})
  if (formModel.lname === "") {
    ElNotification({
      title: '异常',
      type: 'error',
      message: h('i', {style: 'color: #F56C6C'}, '名称不能为空'),
    })
    return
  }
  if (formModel.listen === "") {
    ElNotification({
      title: '异常',
      type: 'error',
      message: h('i', {style: 'color: #F56C6C'}, '监听端口不能为空'),
    })
    return
  }
  if (formModel.saddr === "") {
    ElNotification({
      title: '异常',
      type: 'error',
      message: h('i', {style: 'color: #F56C6C'}, '跳板机IP不能为空'),
    })
    return;
  }
  // 判断用户名是否为空
  if (formModel.user === "") {
    ElNotification({
      title: '异常',
      type: 'error',
      message: h('i', {style: 'color: #F56C6C'}, '跳板机用户名不能为空'),
    })
    return;
  }
  let method = "POST";
  if (formModel.id !== "") {
    method = "PUT"
  }

  sendRequest({
    request:{
      url: "/api/v1/edit.api",
      method: method,
      headers: {'Content-Type': 'application/x-www-form-urlencoded'},
      data: qs.stringify(formModel)
    },
    disableSuccessNotice: true,
    success:(resp:any)=>{
      if (resp.code === 0) {
        ElNotification({
          title: '成功',
          type: 'success',
          message: h('i', {style: 'color: #67C23A'}, '添加成功'),
        })

      } else {
        ElNotification({
          title: '异常',
          type: 'error',
          message: h('i', {style: 'color: #F56C6C'}, resp.msg),
        })
      }
    },
    finally:()=>{
      dialogVisible.form = false;
      loading.close();
    }
  })
}

const selectKey = (data: any) => {
  let input = document.createElement('input');
  input.setAttribute('type', 'file');
  input.setAttribute('accept', '');
  input.value = '';
  input.click();
  input.onchange = function () {
    let reader = new FileReader();
    reader.onload = function (result: any) {
      data.passwd = result.target.result;
    };
    if (input.files && input.files.length > 0) {
      reader.readAsText(input.files[0] || '');
    }

  };
  input.onerror = function (msg) {
    ElNotification({
      title: '异常',
      message: h('i', {style: 'color: error'}, '文件打开失败'),
    })
  }
}


onBeforeMount(() => {
  socket()
})

onBeforeUnmount(() => {
  // 关闭ws
  clearTimeout(setTimeoutId);
  clearInterval(setIntervalId);
  ws.close();
})

const currentLogs = ref<any>({});

const showLogs=(row:any)=>{
  dialogVisibleLogs.value=true;
  currentLogs.value=row;
  logs.splice(0);
  // console.log(logs)
  ws.send("show_"+row['id']);
}

const closeLogs=()=>{
  dialogVisibleLogs.value=false;
  ws.send("close_"+(currentLogs.value['id']));
}



</script>

<template>
  <div class="">
    <el-button type="primary" size="small" :icon="Plus" @click="openForm({type:1,connect:'L',remote:'127.0.0.1:443',listen:'0.0.0.0:443'})">
      添加
    </el-button>
  </div>
  <el-table
      v-loading="dialogVisible.loadding"
      :data="ListData"
      style="width: 100%;margin-top: 20px;font-size: 12px;"
      :highlight-current-row="true"
      :header-cell-style="{background:'#f3f3f3',fontWeight:'bold',color:'#000'}"
      :default-expand-all="true"
      row-key="id"
      :tree-props="{ children: 'son' }">
    <el-table-column prop="lname" label="别名"/>
    <el-table-column label="代理类型">
      <template v-slot="{ row }">
        <el-tag type="primary">{{ row.connectType }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="跳板主机">
      <template v-slot="{ row }">
        <el-tag type="info" class="cur-pointer unselect" @dblclick="handleCopy(row.saddr)">{{ row.saddr }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="user" label="跳板用户"/>
    <el-table-column label="目标地址">
      <template v-slot="{ row }">
        <el-tag type="primary" class="cur-pointer unselect" @dblclick="handleCopy(row.remote)">{{ row.remote }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="listen" label="本端地址">
      <template v-slot="{ row }">
        <el-tag type="danger" class="cur-pointer unselect" @dblclick="handleCopy(row.listen)">{{ row.listen }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="连接状态">
      <template v-slot="{ row }">
        <el-tag :type="row.connectStatus?.type">{{ row.connectStatus?.text }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="操作" width="300">
      <template #default="scope">
        <el-button size="small" text type="primary" @click="openForm({pid:scope.row.id,type:1,connect:'L'})">
          添加
        </el-button>

        <el-button size="small" text type="primary" @click="openForm(scope.row)">编辑</el-button>

        <el-button size="small" text type="info" @click="showLogs(scope.row)">日志</el-button>

        <el-popconfirm confirm-button-text="确认" cancel-button-text="取消" confirm-button-type="danger" title="确认操作？" @confirm="changeConnect(scope.row.id)">
          <template #reference>
            <el-button size="small" text type="warning" @click="">{{ scope.row.active==='Y' ? '暂停' : '连接' }}</el-button>
          </template>
        </el-popconfirm>

        <el-popconfirm
            width="220"
            confirm-button-text="确认"
            confirm-button-type="danger"
            cancel-button-text="取消"
            icon-color="#F56C6C"
            title="是否确认删除"
            @confirm="deleteData( scope.row.id)"
        >
          <template #reference>
            <el-button size="small" text type="danger">删除</el-button>
          </template>
        </el-popconfirm>


      </template>
    </el-table-column>
  </el-table>


  <el-dialog
      :title="dialogVisible.title"
      v-model="dialogVisible.form"
      width="50%" :draggable="true"
      :destroy-on-close="true"
      :close-on-click-modal="false">

    <el-form
        ref="formModelRef"
        :model="formModel"
        label-width="auto"
        :label-position="'left'"
        :show-message="true"
        :scroll-to-error="true"
        size="default"
        autocomplete="on"
        style="padding: 0 20% 0 5%">
      <el-form-item label="名称">
        <el-input v-model="formModel.lname" placeholder="请输入名称" autocomplete="on" @change="formModel.lname=formModel.lname.trim()"/>
        <el-input v-model="formModel.id" type="hidden"/>
        <el-input v-model="formModel.pid" type="hidden"/>
      </el-form-item>
      <el-form-item label="跳板机IP">
        <el-input v-model="formModel.saddr" placeholder="请输入跳板机IP，格式：IP:PORT" autocomplete="on" @change="formModel.saddr=formModel.saddr.trim()"/>
      </el-form-item>
      <el-form-item label="用户名">
        <el-input v-model="formModel.user" placeholder="请输入用户名" autocomplete="on" @change="formModel.user=formModel.user.trim()"/>
      </el-form-item>
      <el-form-item label="密码|密钥">
        <el-radio-group v-model="formModel.type">
          <el-radio :value="1">密码</el-radio>
          <el-radio :value="2">密钥</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="密码">
        <el-input v-model="formModel.passwd" class="input-with-select" type="textarea"
                  placeholder="请输入密码或者密钥"/>
        <el-button size="small" type="primary" style="display: block;position: absolute;right: -62px;top: 0;"
                   @click="selectKey(formModel)">选择
        </el-button>
      </el-form-item>
      <el-form-item label="代理类型">
        <el-radio-group v-model="formModel.connect" @change="connectChange">
          <el-radio v-for="(value,key) in connectTypeMap" :value="key">{{ value }}</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="远程地址" v-show="dialogVisible.remote">
        <el-input v-model="formModel.remote" placeholder="请输入远程地址" autocomplete="on" @change="formModel.remote=formModel.remote.trim()"/>
      </el-form-item>
      <el-form-item label="本地监听地址">
        <el-input v-model="formModel.listen" placeholder="请输入本地监听地址" autocomplete="on" @change="formModel.listen=formModel.listen.trim()"/>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer el-dialog--center">
        <el-button @click="dialogVisible.form = false">取消</el-button>
        <el-button type="primary" @click="submit()">
          保存
        </el-button>
      </div>
    </template>
  </el-dialog>
  <el-dialog title="日志信息" :modal="false" @close="closeLogs" v-model="dialogVisibleLogs" width="50%" height="50%" :draggable="true" :destroy-on-close="true" :close-on-click-modal="false">
    <el-scrollbar>
      <div class="log-viewer">
        <div v-for="(log, index) in logs" :key="index" class="log-line">
          {{ log }}
        </div>
      </div>
    </el-scrollbar>
  </el-dialog>
</template>

<style>
  .log-viewer {
    background-color: #000;
    color: #fff;
    padding: 15px 0;
    font-size: 12px;
  }
  .log-viewer .log-line{
    margin-bottom: 0.5rem;
  }
</style>
