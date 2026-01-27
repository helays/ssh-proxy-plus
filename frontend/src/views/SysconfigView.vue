<script setup lang="ts">
import {onBeforeMount, reactive, ref} from 'vue'
import {ElMessageBox, ElNotification, type FormInstance, type FormRules} from 'element-plus'
import {sendRequest} from "@/components/ts/axios";



const ruleForm = reactive<Record<any, any>>([])

onBeforeMount(() => {
  describeForm()
})

const describeForm = () => {
  sendRequest({
    request: {
      url: '/api/v1/run.sysconfig',
      method: 'get',
    },
    disableSuccessNotice: true,
    success: (data:any) => {
      ruleForm.splice(0, ruleForm.length)
      if (data.code !== 0) {
        ElNotification({
          title: '错误提示',
          message: data.msg,
          type: 'error',
        })
        return
      }
      ruleForm.push(...data.data)
    }
  })
}


const submitForm = () => {
  ElMessageBox.confirm("提示", {
    message: '确认保存？',
    confirmButtonText: '保存',
    cancelButtonText: '取消',
    type: 'info',
    buttonSize: 'default',
    closeOnClickModal: false,
  }).then(() => {
    sendRequest( {
      request: {
        url: '/api/v1/run.sysconfig',
        method: 'POST',
        data: ruleForm,
      },
      success: (data:any) => {
        if (data.code !== 0) {
          ElNotification({
            title: '错误提示',
            message: data.msg,
            type: 'error',
          })
          return
        }
        ElNotification({
          title: '成功提示',
          message: data.msg,
          type: 'success',
        })
      }
    })
  }).catch(() => {
  })
}

</script>

<template>
  <el-form
      ref="ruleFormRef"
      style="max-width: 600px"
      :model="ruleForm"
      status-icon
      label-width="auto"
      class="demo-ruleForm"
      size="default"
  >
    <el-form-item v-for="(item, index) in ruleForm" :prop="item.prop" :label="item.label">
      <component :is="item.name" v-if="item.name=='el-switch'" :active-value="item.component['active-value']" :inactive-value="item.component['inactive-value']"
                 v-model="item.value" :placeholder="item.component.props.placeholder"
                 :autocomplete="item.component.props.autocomplete"/>
      <component :is="item.name" v-else-if="item.type=='text'" :type="item.type"
                 v-model="item.value" :placeholder="item.component.props.placeholder"
                 :autocomplete="item.component.props.autocomplete"/>
      <component :is="item.name" v-else-if="item.type=='password'" :type="item.type"
                 v-model="item.value" :placeholder="item.component.props.placeholder"
                 :autocomplete="item.component.props.autocomplete" :show-password="item.component['show-password']"/>

    </el-form-item>
    <el-form-item>
      <el-button type="primary" @click="submitForm">保存</el-button>
    </el-form-item>
  </el-form>
</template>

<style scoped>

</style>