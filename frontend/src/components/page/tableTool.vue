<script setup lang="ts">
import {onBeforeMount, reactive, ref, watch} from "vue";
import PaginationTools from "@/components/page/paginationTools.vue";
import {listGetHasPage} from "@/components/ts/axios";
import {useRoute} from "vue-router";
import {deepCopyJson, MapVal} from "@/components/ts/utils"
import SearchTool from "@/components/page/searchTool.vue";

const props = defineProps({
  // 扩展table部分
  enableSerial: {type: Boolean, default: true}, // 是否显示序号
  listApi: {type: String, required: false,default:''}, // 列表接口
  requetMethod: {type: String, default: 'GET'}, // 请求方式
  pageField: {type: String, default: 'page'},
  pageSizeField: {type: String, default: 'page_size'},
  searchModel: {type: Object, default: {}},
  refresh: {type: Number},
  enableAutoLoad: {type: Boolean, default: true},
  tableData: {type: Array, default: []},
  afterRequest: {type:Function,},

  // 搜索
  enableSearch: {type: Boolean, default: false},
  searchSize: {type: String, default: 'default'},

  // 自带参数
  style: {
    type: [String, Object],
    default: 'width: 100%;margin-top: 20px;font-size: 12px;font-weight:350',
    required: false
  },
  rowKey: {default: '', type: [String, Function]},
  defaultExpandAll: {type: Boolean, default: true},
  headerCellStyle: {type: [Function, Object], default: {fontWeight: 'bold', color: '#4e4e4e'}},
  highlightCurrentRow: {type: Boolean, default: true},
  dataFormat:{type:Function,required:false}
})


const defaultSearchModel=deepCopyJson(props.searchModel)
const route = useRoute()
const loadding = ref(false)
const objectList = reactive<Record<any, any>[]>(props.tableData as any)
const listTotal = ref(0)

const getLists = (query?: { [key: string]: any }) => {
  loadding.value=true;
  // 检查 query 是否为 undefined 或者是一个空对象
  const actualQuery = query && Object.keys(query).length > 0 ? query : route.query;
  // 循环 defaultSearchModel，如果默认数据里面值不是空的，并切 actualQuery的对应值是空的，就赋值过去。
  for (const key in defaultSearchModel) {
    if (defaultSearchModel[key] !== '' && (actualQuery[key] === undefined) || actualQuery[key] === '') {
      actualQuery[key] = defaultSearchModel[key]
    }
  }
  listGetHasPage(props.listApi, {
    method: props.requetMethod,
    params: actualQuery
  }, listTotal, objectList, {
    loadding:loadding,
    pageField: props.pageField,
    pageSizeField: props.pageSizeField,
    dataFormat:props.dataFormat
  })
}

// 第一次加载
// 组件加载前，需要将query部分反写导model上
const loadFirst=()=>{
  let query={};
  if (props.searchModel){
    // 判断 route.query 是否为空
    if (Object.keys(route.query).length<1){
      // 如果浏览器 query 是空的，说明地址懒重置了，需要重置搜索框
      for (const key in defaultSearchModel) {
        props.searchModel[key] = defaultSearchModel[key]
      }
    }else{
      // 将 route.query 写入到 props.searchModel 中
      for (const key in route.query) {
        if (route.query[key]!==''){
          props.searchModel[key] = route.query[key]
        }
      }
    }
    // 判断搜索框是否为空，为空则删除
    for (const key in props.searchModel){
      if (props.searchModel[key]===''){
        delete props.searchModel[key]
      }
    }
    query=props.searchModel
  }
  getLists(query)
}
// 挂载组件前，调用获取列表的接口
onBeforeMount(() => {
  if (props.enableAutoLoad){
    loadFirst();
  }
})

// 检测到路径变化就刷新列表
watch(() => route.fullPath, (newValue, oldValue) => {
  if (Object.keys(route.query).length<1){
    loadFirst();
    return;
  }
  getLists()
})

watch(objectList,(newValue, oldValue)=>{
  if (props.afterRequest && typeof props.afterRequest==="function"){
    props.afterRequest(objectList)
  }
})

if (props.refresh){
  watch(()=>props.refresh,()=>{
    getLists()
  })
}
</script>

<template>
  <search-tool
      v-if="props.enableSearch"
      v-bind="$attrs"
      :model="props.searchModel"
      :page-field="props.pageField"
      :pageSizeField="props.pageSizeField"
      :default-value="defaultSearchModel"
      :size="props.searchSize">
    <slot name="search"></slot>
  </search-tool>
  <div class="default-top-action-div">
    <slot name="top-action" v-bind="$attrs"/>
  </div>
  <div class="default-table-div">
    <el-table
        v-loading="loadding"
        :data="objectList"
        :style="props.style"
        :row-key="props.rowKey"
        :default-expand-all="props.defaultExpandAll"
        :header-cell-style="props.headerCellStyle"
        :highlight-current-row="props.highlightCurrentRow"
        v-bind="$attrs"
    >
      <el-table-column label="序号" width="70" v-if="props.enableSerial">
        <template #default="scope">
<!--          {{ scope.$index + 1 }}-->
          {{ (MapVal(route.query, props.pageField,1) - 1) * MapVal(route.query,props.pageSizeField,15) + scope.$index + 1 }}
        </template>
      </el-table-column>
      <slot name="default"/>
    </el-table>
  </div>
  <div class="default-bottom-action-div">
    <slot name="bottom-action"/>
  </div>
  <div class="default-pagination-div">
    <paginationTools :total="listTotal" :page-field="props.pageField" :pageSizeField="props.pageSizeField" :search-model="props.searchModel"/>
  </div>
</template>

<style scoped>

</style>