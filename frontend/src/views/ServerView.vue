<script setup lang="ts">

import {Plus, QuestionFilled, Refresh} from "@element-plus/icons-vue";
import {onBeforeMount, onBeforeUnmount, reactive, ref} from 'vue'
import {ElNotification, type FormInstance, type FormRules} from "element-plus";
import {ValidatePassword,GeneratePassword} from "@/components/PassowordTools.vue"
import {CatchMessageNotification,ThenErrorMsgNotification} from "@/components/RespTools.vue"
import {TxtToClipboard} from "@/components/Copy2ClipboardTools.vue"
import {sendRequest} from "@/components/ts/axios";

let autoLoadOrder:any;

onBeforeMount(() => {
  getDescribeRegions(getOrderLists)
})

onBeforeUnmount(()=>{
  clearInterval(autoLoadOrder)
})

const orderList = reactive<Record<any, any>[]>([])

// 获取订单列表
const getOrderLists=()=>{
  clearInterval(autoLoadOrder)

  sendRequest({
    request:{
      url: '/api/v1/ali.describe.instances',
      method: 'get',
    },
    disableSuccessNotice:true,
    success: (data:any) => {
      orderList.splice(0,orderList.length)
      if (data.code !== 0) {
        ThenErrorMsgNotification(data)
        return
      }
      orderList.push(...data.data)
    },
    finally: () => {
      autoLoadOrder=setTimeout(()=>{
        getOrderLists()
      },30*1000)
    }
  })
}

const openAliEcsDialogVisible = ref(false)

const aliEcsForm = reactive({
  instance_charge_type: 'PostPaid',
  region_id: 'cn-hongkong',
  image_id: 'centos_7_9_x64_20G_alibase_20240628.vhd',
  instance_type: 'ecs.t5-lc2m1.nano',
  password_inherit: false,
  auto_renew: false,
  auto_release_time: 1,
  auto_pay: true,
  internet_charge_type: 'PayByBandwidth',
  internet_max_bandwidth: 1,
  dry_run: true,
  security_group_id: '',
  password: GeneratePassword(30),
  system_disk: {
    category: 'cloud_efficiency',
    size: 20,
  },
  io_optimized: 'optimized',
  security_enhancement_strategy: 'Active',
  v_switch_id: '',
  local_listen_addr:'0.0.0.0:1080',
});
const ruleFormRef = ref<FormInstance>()

const rules = reactive<FormRules<typeof aliEcsForm>>({
  password: [{ validator: (rule: any, value: any, callback: any)=>{
    if (value === '') {
      callback(new Error('请输入密码'))
    }else if(!ValidatePassword(value)){
      callback(new Error('实例的密码。长度为 8 至 30 个字符，必须同时包含大小写英文字母、数字和特殊符号中的三类字符'))
    } else {
      callback()
    }
    }, trigger: 'blur' }],
  local_listen_addr: [{ validator: (rule: any, value: any, callback: any)=>{
      if (value === '') {
        callback(new Error('请输入本地监听地址'))
      }else{
        callback()
      }
    }, trigger: 'blur' }]
})

// 申请创建ECS
const createEcs = (formEl: FormInstance | undefined) => {
  if (!formEl) return
  formEl.validate((valid) => {
    if(!valid){
      return;
    }
    sendRequest({
      request:{
        url: '/api/v1/ali.run.instances',
        method: 'post',
        data: aliEcsForm
      },
      success: (data:any) => {
        // 判断 data.data.Message 是否包含 Request validation has been passed with DryRun flag set
        if (data.code !== 0) {
          ThenErrorMsgNotification(data)
          return;
        }
        openAliEcsDialogVisible.value = false

        getOrderLists()
        ElNotification({
          title: '成功提示',
          message: '添加成功，请稍后手动刷新页面查看开通情况',
          type: 'success',
        })
      },
      catch: (err:any) => {
        getOrderLists()
        if (err.response.data.data.Message.indexOf('Request validation has been passed with DryRun flag set')===0){
          openAliEcsDialogVisible.value = false
          ElNotification({
            title: '成功提示',
            message: '测试通过',
            type: 'success',
          })
          return;
        }
        CatchMessageNotification(err)
      },
    })

  })
}

const regionList = reactive<Record<any, any>[]>([])
// 获取地域列表
const getDescribeRegions = (...call:any | undefined) => {
  sendRequest({
    request:{
      url: '/api/v1/describe.regions',
      method: 'get',
      params: {
        instance_charge_type: aliEcsForm.instance_charge_type,
        resource_type: 'instance',
        accept_language: 'zh-CN'
      }
    },
    disableSuccessNotice:true,
    success: (data:any) => {
      regionList.splice(0, regionList.length)
      if (data.code !== 0) {
        ThenErrorMsgNotification(data)
        return;
      }
      regionList.push(...data.msg.Region)
      changeRegion(aliEcsForm.region_id)
      if (call && call.length>0){
        call[0]()
      }
    },
  })
}

const availableResourceList = reactive<Record<any, any>[]>([])
// 查询某一可用区的资源列表
const getDescribeAvailableResource = () => {

  sendRequest({
    request:{
      url: '/api/v1/describe.available.resource',
      method: 'get',
      params: {
        region_id: aliEcsForm.region_id,
        instance_charge_type: aliEcsForm.instance_charge_type,
        destination_resource: 'InstanceType',
      }
    },
    disableSuccessNotice:true,
    success: (data:any) => {
      availableResourceList.splice(0, availableResourceList.length)
      if (data.code !== 0) {
        ThenErrorMsgNotification(data)
        return
      }
      availableResourceList.push(...data.msg.AvailableZone)
    }
  })
}

const securityGroupsList=reactive<Record<any,any>>([])

// 获取安全组列表
const getDescribeSecurityGroups=()=>{
  sendRequest({
    request:{
      url: '/api/v1/describe.security.groups',
      method: 'get',
      params: {region_id: aliEcsForm.region_id}
    },
    disableSuccessNotice:true,
    success: (data:any) => {
      securityGroupsList.splice(0,securityGroupsList.length)
      if (data.code !== 0) {
        ThenErrorMsgNotification(data)
        return
      }
      securityGroupsList.push(...data.msg.SecurityGroup)
      if (data.msg.SecurityGroup.length===1){
        aliEcsForm.security_group_id=data.msg.SecurityGroup[0].SecurityGroupId
      }
    }
  })
}

const vSwitchesList=reactive<Record<any,any>>([])
const getDescribeVSwitches=()=>{
  sendRequest({
    request:{
      url: '/api/v1/describe.v.switches',
      method: 'get',
      params: {region_id: aliEcsForm.region_id}
    },
    disableSuccessNotice:true,
    success: (data:any) => {
      vSwitchesList.splice(0,vSwitchesList.length)
      if (data.code !== 0) {
        ThenErrorMsgNotification(data)
        return
      }
      vSwitchesList.push(...data.msg.VSwitch)
      if (data.msg.VSwitch.length===1){
        aliEcsForm.v_switch_id=data.msg.VSwitch[0].VSwitchId
      }
    }
  })

}

// 改变地域的时候，联动资源规格
const changeRegion = (value: any) => {
  getDescribeSecurityGroups()
  getDescribeVSwitches()
}

const deleteInstance=(id:number)=>{

  sendRequest({
    request:{
      url: '/api/v1/ali.del.instances',
      method: 'post',
      params: {id: id}
    },
    disableSuccessNotice:true,
    success: (data:any) => {
      if (data.code!==0){
        ThenErrorMsgNotification(data)
        return
      }
      getOrderLists()
      ElNotification({
        title: '成功提示',
        dangerouslyUseHTMLString:true,
        message: '<p>'+data.msg+'</p><p>请求ID:'+data.data+'</p>',
        type: 'success',
      })
    }
  })
}

</script>

<template>
  <div class="">
    <el-button type="primary" size="default" :icon="Plus" @click="openAliEcsDialogVisible=true">
      开通ECS
    </el-button>
  </div>
  <el-dialog
      v-model="openAliEcsDialogVisible"
      title="申请开通阿里云ECS"
      width="55%"
      :close-on-click-modal="false"
      draggable
      align-center>
    <el-form :model="aliEcsForm" ref="ruleFormRef" :rules="rules" label-width="auto" size="default" style="width: 90%" label-position="left">

      <el-row style="justify-content: space-between">
        <el-col :md="24" :lg="11">
          <el-form-item label="本地监听地址">
            <el-input v-model="aliEcsForm.local_listen_addr" placeholder="本地监听地址"/>
          </el-form-item>
        </el-col>
        <el-col :md="24" :lg="11">
          <el-form-item label="计费方式">
            <el-radio-group v-model="aliEcsForm.instance_charge_type" @change="getDescribeRegions">
              <el-radio value="PostPaid">按量计费</el-radio>
              <el-radio value="PrePaid">包年包月</el-radio>
            </el-radio-group>
          </el-form-item>
        </el-col>
      </el-row>


      <el-row style="justify-content: space-between">
        <el-col :md="24" :lg="11">
          <el-form-item label="地域">
            <el-select v-model="aliEcsForm.region_id" filterable placeholder="选择实例所属地域" @change="changeRegion">
              <el-option v-for="(item,index) in regionList" :label="item.LocalName" :value="item.RegionId"/>
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :md="24" :lg="11">
          <el-form-item label="资源规格">
            <el-input v-model="aliEcsForm.instance_type" placeholder="选择资源规格"/>
          </el-form-item>
        </el-col>
      </el-row>



      <el-form-item label="镜像">
        <el-input v-model="aliEcsForm.image_id" placeholder="选择镜像"/>
      </el-form-item>

      <el-form-item label="使用预设密码">
        <el-switch v-model="aliEcsForm.password_inherit"/>
      </el-form-item>
      <el-form-item label="实例密码" v-if="!aliEcsForm.password_inherit" :inline-message="true" prop="password">
        <el-input v-model="aliEcsForm.password" placeholder="输入实例密码" type="password" show-password>
          <template #prefix>
            <el-tooltip placement="auto" content="实例的密码。长度为 8 至 30 个字符，必须同时包含大小写英文字母、数字和特殊符号中的三类字符">
              <el-icon style="cursor:default;"><QuestionFilled/></el-icon>
            </el-tooltip>
          </template>
          <template #suffix>
            <el-icon style="cursor: pointer;" @click="aliEcsForm.password=GeneratePassword(8)"><Refresh /></el-icon>
          </template>
        </el-input>
      </el-form-item>



      <el-form-item label="自动释放时间">
        <el-input v-model="aliEcsForm.auto_release_time" type="number">
          <template #append>小时</template>
        </el-input>
      </el-form-item>

      <el-form-item label="网络计费">
        <el-radio-group v-model="aliEcsForm.internet_charge_type">
          <el-radio value="PayByBandwidth">固定带宽</el-radio>
          <el-radio value="PayByTraffic">使用流量</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item label="带宽" v-if="aliEcsForm.internet_charge_type=='PayByBandwidth'">
        <el-input v-model="aliEcsForm.internet_max_bandwidth" type="number">
          <template #append>Mb</template>
        </el-input>
      </el-form-item>
      <el-form-item label="安全组">
        <el-select v-model="aliEcsForm.security_group_id" filterable placeholder="选择安全组">
          <el-option v-for="(item,index) in securityGroupsList" :label="item.SecurityGroupName" :value="item.SecurityGroupId"/>
        </el-select>
      </el-form-item>

      <el-form-item label="交换机">
        <el-select v-model="aliEcsForm.v_switch_id" filterable placeholder="选择交换机">
          <el-option v-for="(item,index) in vSwitchesList" :label="item.VSwitchName ? item.VSwitchName : item.VSwitchId" :value="item.VSwitchId"/>
        </el-select>
      </el-form-item>

      <el-row style="justify-content: space-between">
        <el-col :md="24" :lg="12">
          <el-form-item label="I/O优化">
            <el-radio-group v-model="aliEcsForm.io_optimized">
              <el-radio value="none">不启用</el-radio>
              <el-radio value="optimized">启用</el-radio>
            </el-radio-group>
          </el-form-item>
        </el-col>
        <el-col :md="24" :lg="12">
          <el-form-item label="安全加固">
            <el-radio-group v-model="aliEcsForm.security_enhancement_strategy">
              <el-radio value="Active">启用</el-radio>
              <el-radio value="Deactive">不启用</el-radio>
            </el-radio-group>
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="系统盘">
        <el-radio-group v-model="aliEcsForm.system_disk.category">
          <el-radio value="cloud">普通云盘</el-radio>
          <el-radio value="cloud_efficiency">高效云盘</el-radio>
          <el-radio value="cloud_essd">ESSD云盘</el-radio>
          <el-radio value="cloud_ssd">SSD云盘</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="系统盘大小">
        <el-input v-model="aliEcsForm.system_disk.size" type="number">
          <template #append>GB</template>
        </el-input>
      </el-form-item>

      <el-row style="justify-content: space-between">
        <el-form-item label="自动续费">
          <el-switch v-model="aliEcsForm.auto_renew"/>
        </el-form-item>
        <el-form-item label="自动支付">
          <el-switch v-model="aliEcsForm.auto_pay"/>
        </el-form-item>
        <el-form-item label="测试模式">
          <el-switch v-model="aliEcsForm.dry_run"/>
        </el-form-item>
      </el-row>


    </el-form>


    <template #footer>
      <div class="dialog-footer">
        <el-button @click="openAliEcsDialogVisible = false" size="default">取消</el-button>
        <el-button type="primary" @click="createEcs(ruleFormRef)" size="default">
          申请开通
        </el-button>
      </div>
    </template>
  </el-dialog>
  <div style="margin-top: 20px"></div>
  <el-table :data="orderList" style="width: 100%">
    <el-table-column label="序号" width="80" prop="id" />
    <el-table-column label="实例ID" show-overflow-tooltip>
      <template #default="scope">
        <el-text class="not-select" @dblclick="TxtToClipboard(scope.row.instance_id)">{{ scope.row.instance_id }}</el-text>
      </template>
    </el-table-column>
    <el-table-column label="地区" prop="region_id" />
    <el-table-column label="资源类型" prop="instance_type" />
    <el-table-column label="付费类型">
      <template #default="scope">
        {{ scope.row.instance_charge_type =='PostPaid' ? '按量付费' : '包年包月' }}
      </template>
    </el-table-column>
    <el-table-column label="带宽类型">
      <template #default="scope">
        {{ scope.row.internet_charge_type =='PayByBandwidth' ? '固定带宽：'+scope.row.internet_max_bandwidth+"Mb" : '流量' }}
      </template>
    </el-table-column>
    <el-table-column label="监听地址" prop="local_listen_addr" />
    <el-table-column label="实例IP" prop="public_ip_address" />
    <el-table-column label="保留时长">
      <template #default="scope">
        {{ scope.row.auto_release_time+"小时" }}
      </template>
    </el-table-column>
    <el-table-column label="自动支付">
      <template #default="scope">
        <el-text class="mx-1" :type="scope.row.auto_pay? 'danger' : 'primary'">{{ scope.row.auto_pay ? '是' : '否' }}</el-text>
      </template>
    </el-table-column>
    <el-table-column label="自动续费">
      <template #default="scope">
        <el-text class="mx-1" :type="scope.row.auto_renew ? 'danger' : 'primary'">{{ scope.row.auto_renew ? '是' : '否' }}</el-text>
      </template>
    </el-table-column>

    <el-table-column label="运行状态">
      <template #default="scope">
        <el-tooltip raw-content>
            <template #default>
              <el-text class="mx-1" :type="scope.row.run_status==='Running' ? 'primary' : 'danger'">{{ scope.row.run_status==='' ? '-': scope.row.run_status }}</el-text>
            </template>
          <template #content>
            <p>订购状态码：{{ scope.row.order_status }}</p>
            <p>订购错误信息：{{ scope.row.err_message }}</p>
            <p>订购错误数据：<pre>{{ scope.row.err_data }}</pre></p>
            <p>查询状态码：{{ scope.row.query_status }}</p>
            <p>查询错误信息：{{ scope.row.query_err_message }}</p>
            <p>查询错误数据：<pre>{{ scope.row.query_err_data }}</pre></p>
          </template>
        </el-tooltip>


      </template>
    </el-table-column>

    <el-table-column label="测试模式">
      <template #default="scope">
        <el-text class="mx-1" :type="scope.row.dry_run? 'danger' : 'primary'">{{ scope.row.dry_run ? '是' : '否' }}</el-text>
      </template>
    </el-table-column>
    <el-table-column label="创建时间" prop="create_time" />
    <el-table-column label="操作">
      <template #default="scope">
        <el-popconfirm confirm-button-text="确认" cancel-button-text="取消" confirm-button-type="danger" title="确认删除？" @confirm="deleteInstance(scope.row.id)">
          <template #reference>
            <el-button size="small" type="danger">删除</el-button>
          </template>
        </el-popconfirm>
      </template>
    </el-table-column>
  </el-table>
</template>

<style scoped>

</style>