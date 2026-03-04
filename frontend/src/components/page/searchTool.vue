<script setup lang="ts">
import {useRoute,useRouter} from "vue-router";
import {deepCopyJson} from "@/components/ts/utils";

const props=defineProps({
  model: {type: Object, default:{}},
  defaultValue:{type:Object,default:{}},
  pageField: {type: String, default: 'page'},
  pageSizeField: {type: String, default: 'page_size'},
})

const router=useRouter()
let route=useRoute()

// 提交搜索
const onSubmit=()=>{
  // 需要自动去除每个字段前后的空格
  for(let key in props.model){
    // 这里需要容错，有trim函数的才处理
    if(props.model[key] && props.model[key].trim){
      props.model[key]=props.model[key].trim()
    }
    // 如果值是空的，就删除
    if(props.model[key]==''){
      delete props.model[key]
    }
    delete props.model[props.pageField]
    delete props.model[props.pageSizeField]
  }
  router.push({
    path: route.path,
    query: props.model
  })
}

// 重置
const resetModel=()=>{
  for(let key in props.defaultValue){
    props.model[key]=props.defaultValue[key]
  }
}
</script>

<template>
  <div class="default-search-div">
    <el-form :inline="true" :model="props.model" class="demo-form-inline" v-bind="$attrs">
      <slot name="default"/>
      <el-form-item>
        <el-button type="primary" @click="onSubmit" size="small">查询</el-button>
      </el-form-item>
      <el-form-item>
        <el-button text @click="resetModel" size="small">重置</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<style scoped>

</style>